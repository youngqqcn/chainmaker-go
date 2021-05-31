/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	"sync"

	commonPb "chainmaker.org/chainmaker/pb-go/common"
	"chainmaker.org/chainmaker-go/protocol"
	"chainmaker.org/chainmaker-go/utils"
)

type ManagerDelegate struct {
	lock            sync.Mutex
	blockchainStore protocol.BlockchainStore
}

func (m *ManagerDelegate) calcSnapshotFingerPrint(snapshot *SnapshotImpl) utils.BlockFingerPrint {
	if snapshot == nil {
		return ""
	}
	chainId := snapshot.chainId
	blockHeight := snapshot.blockHeight
	blockTimestamp := snapshot.blockTimestamp
	blockProposer := snapshot.blockProposer
	preBlockHash := snapshot.preBlockHash

	return utils.CalcFingerPrint(chainId, blockHeight, blockTimestamp, blockProposer, preBlockHash)
}

func (m *ManagerDelegate) makeSnapshotImpl(block *commonPb.Block, blockHeight int64) *SnapshotImpl {
	// If the corresponding Snapshot does not exist, create one
	txCount := len(block.Txs) // as map init size
	snapshotImpl := &SnapshotImpl{
		blockchainStore: m.blockchainStore,
		sealed:          false,
		preSnapshot:     nil,

		txResultMap: make(map[string]*commonPb.Result, txCount),

		chainId:        block.Header.ChainId,
		blockHeight:    block.Header.BlockHeight,
		blockTimestamp: block.Header.BlockTimestamp,
		blockProposer:  block.Header.Proposer,
		preBlockHash:   block.Header.PreBlockHash,

		txTable:    nil,
		readTable:  make(map[string]*sv, txCount),
		writeTable: make(map[string]*sv, txCount),
	}
	return snapshotImpl
}
