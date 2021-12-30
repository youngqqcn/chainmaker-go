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

	require.NoError(t, routine.addTask(NodeStatusMsg{from: "node1"}))
	require.NoError(t, routine.addTask(NodeStatusMsg{from: "node2"}))
	require.NoError(t, routine.addTask(NodeStatusMsg{from: "node3"}))

	for i := range mock.receiveItems {
		item := mock.receiveItems[i].(NodeStatusMsg)
		require.EqualValues(t, fmt.Sprintf("node%d", i+1), item.from)
	}
	routine.end()
}

func TestPriority(t *testing.T) {
	//var ok bool
	var err error
	//var item []queue.Item
	q := queue.NewPriorityQueue(bufferSize, true)
	err = q.Put(EqualLevel{})
	require.NoError(t, err)
	err = q.Put(SyncedBlockMsg{msg: []byte("msg"), from: "node1"})
	require.NoError(t, err)
	err = q.Put(NodeStatusMsg{msg: syncPb.BlockHeightBCM{BlockHeight: 1}, from: "node1"})
	require.NoError(t, err)
	err = q.Put(SchedulerMsg{})
	require.NoError(t, err)
	err = q.Put(LivenessMsg{})
	require.NoError(t, err)
	err = q.Put(ReceivedBlocks{
		blks: []*commonPb.Block{
			{Header: &commonPb.BlockHeader{BlockHeight: 9}},
			{Header: &commonPb.BlockHeader{BlockHeight: 11}},
			{Header: &commonPb.BlockHeader{BlockHeight: 12}},
		}, from: "node1"})
	require.NoError(t, err)
	err = q.Put(ReceivedBlocksWithRwSets{})
	require.NoError(t, err)
	err = q.Put(ProcessBlockMsg{})
	require.NoError(t, err)
	err = q.Put(ProcessBlockWithRwSetMsg{})
	require.NoError(t, err)
	err = q.Put(DataDetection{})
	require.NoError(t, err)
	err = q.Put(ProcessedBlockResp{})
	require.NoError(t, err)

	require.Equal(t, q.Len(), 11)
	//item, err = q.Get(1)
	//require.NoError(t, err)
	//_, ok = item[0].(NodeStatusMsg)
	//require.Equal(t, ok, true)
	//item, err = q.Get(1)
	//_, ok = item[0].(SyncedBlockMsg)
	//require.Equal(t, ok, true)
	//item, err = q.Get(1)
	//_, ok = item[0].(ReceivedBlocks)
	//require.Equal(t, ok, true)
}
