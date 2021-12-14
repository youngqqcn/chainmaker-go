/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// TxScheduler evidence transaction scheduler structure
type TxSchedulerEvidence struct {
	delegate *TxScheduler
}

func (ts *TxSchedulerEvidence) Schedule(block *commonPb.Block, txBatch []*commonPb.Transaction,
	snapshot protocol.Snapshot) (map[string]*commonPb.TxRWSet, map[string][]*commonPb.ContractEvent, error) {
	return ts.delegate.Schedule(block, txBatch, snapshot)
}

// SimulateWithDag based on the dag in the block, perform scheduling and execution evidence transactions
func (ts *TxSchedulerEvidence) SimulateWithDag(block *commonPb.Block,
	snapshot protocol.Snapshot) (map[string]*commonPb.TxRWSet, map[string]*commonPb.Result, error) {
	return ts.delegate.SimulateWithDag(block, snapshot)
}

func (ts *TxSchedulerEvidence) Halt() {
	ts.delegate.Halt()
}
