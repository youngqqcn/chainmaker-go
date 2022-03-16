/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"fmt"
	"testing"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"

	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"

	"chainmaker.org/chainmaker/protocol/v2/test"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/stretchr/testify/require"
)

type MockHandler struct {
	receiveItems []queue.Item
}

func NewMockHandler() *MockHandler {
	return &MockHandler{receiveItems: make([]queue.Item, 0, 10)}
}

func (mock *MockHandler) handler(item queue.Item) (queue.Item, error) {
	mock.receiveItems = append(mock.receiveItems, item)
	return nil, nil
}
func (mock *MockHandler) getState() string {
	return ""
}

func TestAddTask(t *testing.T) {
	mock := NewMockHandler()
	routine := NewRoutine("mock", mock.handler, mock.getState, &test.GoLogger{})
	require.NoError(t, routine.begin())

	require.NoError(t, routine.addTask(&NodeStatusMsg{from: "node1"}))
	require.NoError(t, routine.addTask(&NodeStatusMsg{from: "node2"}))
	require.NoError(t, routine.addTask(&NodeStatusMsg{from: "node3"}))

	for i := range mock.receiveItems {
		item := mock.receiveItems[i].(*NodeStatusMsg)
		require.EqualValues(t, fmt.Sprintf("node%d", i+1), item.from)
	}
	routine.end()
}

func TestPriority(t *testing.T) {
	q := queue.NewPriorityQueue(bufferSize, true)
	require.NoError(t, q.Put(&SyncedBlockMsg{msg: []byte("msg"), from: "node1"}))
	require.NoError(t, q.Put(&NodeStatusMsg{msg: syncPb.BlockHeightBCM{BlockHeight: 1}, from: "node1"}))
	require.NoError(t, q.Put(&SchedulerMsg{}))
	require.NoError(t, q.Put(&LivenessMsg{}))
	err := q.Put(&ReceivedBlockInfos{
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
	require.NoError(t, err)
	require.NoError(t, q.Put(&ProcessBlockMsg{}))
	require.NoError(t, q.Put(&DataDetection{}))
	require.NoError(t, q.Put(&ProcessedBlockResp{}))

	require.Equal(t, q.Len(), 8)
	item, err := q.Get(1)
	require.NoError(t, err)
	_, ok := item[0].(*ReceivedBlockInfos)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*ProcessedBlockResp)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*ProcessBlockMsg)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*SyncedBlockMsg)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*NodeStatusMsg)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*LivenessMsg)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*DataDetection)
	require.Equal(t, ok, true)
	item, err = q.Get(1)
	require.NoError(t, err)
	_, ok = item[0].(*SchedulerMsg)
	require.Equal(t, ok, true)
	require.Equal(t, q.Len(), 0)
}
