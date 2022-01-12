/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"
	"github.com/Workiva/go-datastructures/queue"
)

const (
	priorityTop    = 3
	priorityMiddle = 2
	priorityLow    = 1
	//priorityBase   = 0
)

type Priority interface {
	Level() int
}

func Compare(parent, other queue.Item) int {
	//doProcessBlockTk 	-> ... priority first
	//doScheduleTk 		-> ... priority second
	//doNodeStatusTk 	-> ... priority third
	var (
		ok   bool
		p, o Priority
	)
	if p, ok = parent.(Priority); !ok {
		return 0
	}
	if o, ok = other.(Priority); !ok {
		return 0
	}
	if p.Level() < o.Level() {
		return 1
	} else if p.Level() == o.Level() {
		return 0
	}
	return -1
}

type SyncedBlockMsg struct {
	msg  []byte
	from string
}

func (m *SyncedBlockMsg) Level() int {
	return priorityMiddle
}

func (m *SyncedBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type NodeStatusMsg struct {
	msg  syncPb.BlockHeightBCM
	from string
}

func (m *NodeStatusMsg) Level() int {
	return priorityLow
}

func (m *NodeStatusMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type SchedulerMsg struct{}

func (m *SchedulerMsg) Level() int {
	return priorityLow
}

func (m *SchedulerMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type LivenessMsg struct{}

func (m *LivenessMsg) Level() int {
	return priorityLow
}

func (m *LivenessMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type ReceivedBlockInfos struct {
	*syncPb.SyncBlockBatch
	from string
}

func (m *ReceivedBlockInfos) Level() int {
	return priorityTop
}

func (m *ReceivedBlockInfos) Compare(other queue.Item) int {
	return Compare(m, other)
}

// processor events

type ProcessBlockMsg struct{}

func (m *ProcessBlockMsg) Level() int {
	return priorityTop
}

func (m *ProcessBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type DataDetection struct{}

func (m *DataDetection) Level() int {
	return priorityLow
}

func (m *DataDetection) Compare(other queue.Item) int {
	return Compare(m, other)
}

type processedBlockStatus int64

const (
	ok processedBlockStatus = iota
	dbErr
	addErr
	hasProcessed
	validateFailed
)

type ProcessedBlockResp struct {
	height uint64
	status processedBlockStatus
	from   string
}

func (m *ProcessedBlockResp) Level() int {
	return priorityTop
}

func (m *ProcessedBlockResp) Compare(other queue.Item) int {
	return Compare(m, other)
}
