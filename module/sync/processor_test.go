/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"testing"

	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2/test"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestProcessorReceivedBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ledger := newMockLedgerCache(ctrl, &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 100}})
	mockVerifier := NewMockVerifyAndCommit(ledger)
	processor := newProcessor(mockVerifier, ledger, &test.GoLogger{})

	// 1. Receive the blocks which has been confirmed
	result, err := processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 10}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 11}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 12}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})
	require.Nil(t, result)
	require.NoError(t, err)
	require.EqualValues(t, 0, len(processor.queue))

	// 2. Receive the blocks which part of has been confirmed
	result, err = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 101}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 102}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})
	require.Nil(t, result)
	require.NoError(t, err)
	require.EqualValues(t, 2, len(processor.queue))

	// 3. Receive the blocks which not been confirmed
	result, err = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 103}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 104}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 105}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})
	require.Nil(t, result)
	require.NoError(t, err)
	require.EqualValues(t, 5, len(processor.queue))

	// 4. Repeat receive the blocks which not been confirmed
	result, err = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 103}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 104}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 105}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})
	require.Nil(t, result)
	require.NoError(t, err)
	require.EqualValues(t, 5, len(processor.queue))
}

func TestProcessorProcessBlockMsg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ledger := newMockLedgerCache(ctrl, &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 100}})
	mockVerifier := NewMockVerifyAndCommit(ledger)
	processor := newProcessor(mockVerifier, ledger, &test.GoLogger{})

	// 1. Receive the blocks which has not been confirmed
	_, _ = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 102}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 103}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 104}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})

	// 2. process block, but not have the block which the height is 101
	ret, err := processor.handler(&ProcessBlockMsg{})
	require.Nil(t, ret)
	require.NoError(t, err)
	require.EqualValues(t, 3, len(processor.queue))

	// 3. Add the block which height is 101
	_, _ = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 101}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})
	ret, err = processor.handler(&ProcessBlockMsg{})
	require.NoError(t, err)
	require.EqualValues(t, ok, ret.(*ProcessedBlockResp).status)
	require.EqualValues(t, 3, len(processor.queue))
	require.EqualValues(t, 1, len(mockVerifier.receiveItem))
	require.EqualValues(t, 101, ledger.GetLastCommittedBlock().Header.BlockHeight)

	// 4. process next blocks
	for i := 1; i <= 3; i++ {
		ret, err = processor.handler(&ProcessBlockMsg{})
		require.NoError(t, err)
		require.EqualValues(t, 3-i, len(processor.queue))
		require.EqualValues(t, ok, ret.(*ProcessedBlockResp).status)
		require.EqualValues(t, 1+i, len(mockVerifier.receiveItem))
		require.EqualValues(t, 101+i, ledger.GetLastCommittedBlock().Header.BlockHeight)
	}
	require.EqualValues(t, 4, processor.hasProcessedBlock())
}

func TestDataDetection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ledger := newMockLedgerCache(ctrl, &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 100}})
	mockVerifier := NewMockVerifyAndCommit(ledger)
	processor := newProcessor(mockVerifier, ledger, &test.GoLogger{})

	// 1. Receive the blocks which has not been confirmed
	_, _ = processor.handler(&ReceivedBlockInfos{
		SyncBlockBatch: &syncPb.SyncBlockBatch{
			Data: &syncPb.SyncBlockBatch_BlockinfoBatch{
				BlockinfoBatch: &syncPb.BlockInfoBatch{
					Batch: []*commonPb.BlockInfo{
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 102}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 103}}},
						{Block: &commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 104}}},
					},
				},
			},
			WithRwset: false,
		},
		from: "node1",
	})

	// 2. no blocks will be deleted
	ret, err := processor.handler(&DataDetection{})
	require.Nil(t, ret)
	require.NoError(t, err)
	require.EqualValues(t, 3, len(processor.queue))

	// 3. modify ledger status and trigger data detection
	ledger.SetLastCommittedBlock(&commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 102}})
	ret, err = processor.handler(&DataDetection{})
	require.Nil(t, ret)
	require.NoError(t, err)
	require.EqualValues(t, 2, len(processor.queue))

	// 4. modify ledger status and trigger data detection
	ledger.SetLastCommittedBlock(&commonPb.Block{Header: &commonPb.BlockHeader{BlockHeight: 120}})
	ret, err = processor.handler(&DataDetection{})
	require.Nil(t, ret)
	require.NoError(t, err)
	require.EqualValues(t, 0, len(processor.queue))
}
