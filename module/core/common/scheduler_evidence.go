/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	commonpb "chainmaker.org/chainmaker-go/pb/protogo/common"
	"chainmaker.org/chainmaker-go/protocol"
)

// TxScheduler evidence transaction scheduler structure
type TxSchedulerEvidence struct {
	delegate *TxScheduler
}

func (ts *TxSchedulerEvidence) Schedule(block *commonpb.Block, txBatch []*commonpb.Transaction, snapshot protocol.Snapshot) (map[string]*commonpb.TxRWSet, map[string][]*commonpb.ContractEvent, error) {
	return ts.delegate.Schedule(block, txBatch, snapshot)
}

// SimulateWithDag based on the dag in the block, perform scheduling and execution evidence transactions
func (ts *TxSchedulerEvidence) SimulateWithDag(block *commonpb.Block, snapshot protocol.Snapshot) (map[string]*commonpb.TxRWSet, map[string]*commonpb.Result, error) {
	return ts.delegate.SimulateWithDag(block, snapshot)
}

func (ts *TxSchedulerEvidence) Halt() {
	ts.delegate.Halt()
}
