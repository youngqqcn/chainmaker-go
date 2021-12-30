/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"
	"github.com/Workiva/go-datastructures/queue"
)

const (
	priorityTop    = 3
	priorityMiddle = 2
	priorityLow    = 1
	priorityBase   = 0
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
	}
	return 0
}

type EqualLevel struct{}

func (e EqualLevel) Level() int {
	return priorityBase
}

func (e EqualLevel) Compare(other queue.Item) int {
	return Compare(e, other)
}

type SyncedBlockMsg struct {
	EqualLevel
	msg  []byte
	from string
}

func (m SyncedBlockMsg) Level() int {
	return priorityMiddle
}

func (m SyncedBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type NodeStatusMsg struct {
	EqualLevel
	msg  syncPb.BlockHeightBCM
	from string
}

func (m NodeStatusMsg) Level() int {
	return priorityLow
}

func (m NodeStatusMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type SchedulerMsg struct {
	EqualLevel
}

func (m SchedulerMsg) Level() int {
	return priorityLow
}

func (m SchedulerMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type LivenessMsg struct {
	EqualLevel
}

func (m LivenessMsg) Level() int {
	return priorityLow
}

func (m LivenessMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type ReceivedBlocks struct {
	blks []*commonPb.Block
	from string
	EqualLevel
}

func (m ReceivedBlocks) Level() int {
	return priorityTop
}

func (m ReceivedBlocks) Compare(other queue.Item) int {
	return Compare(m, other)
}

type ReceivedBlocksWithRwSets struct {
	blkinfos []*commonPb.BlockInfo
	from     string
	EqualLevel
}

// processor events

type ProcessBlockMsg struct {
	EqualLevel
}

func (m ProcessBlockMsg) Level() int {
	return priorityTop
}

func (m ProcessBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

type ProcessBlockWithRwSetMsg struct {
	EqualLevel
}

type DataDetection struct {
	EqualLevel
}

func (m DataDetection) Level() int {
	return priorityLow
}

func (m DataDetection) Compare(other queue.Item) int {
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
	EqualLevel
}

func (m ProcessedBlockResp) Level() int {
	return priorityTop
}

func (m ProcessedBlockResp) Compare(other queue.Item) int {
	return Compare(m, other)
}
