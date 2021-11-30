/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	"errors"
	"math"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

type SnapshotEvidence struct {
	delegate *SnapshotImpl
	log      protocol.Logger
}

func (s *SnapshotEvidence) GetPreSnapshot() protocol.Snapshot {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetPreSnapshot()
}

func (s *SnapshotEvidence) SetPreSnapshot(snapshot protocol.Snapshot) {
	if s.delegate == nil {
		return
	}
	s.delegate.SetPreSnapshot(snapshot)
}

func (s *SnapshotEvidence) GetBlockchainStore() protocol.BlockchainStore {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetBlockchainStore()
}

func (s *SnapshotEvidence) GetSnapshotSize() int {
	if s.delegate == nil {
		return -1
	}
	return s.delegate.GetSnapshotSize()
}

func (s *SnapshotEvidence) GetTxTable() []*commonPb.Transaction {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetTxTable()
}

func (s *SnapshotEvidence) GetSpecialTxTable() []*commonPb.Transaction {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetSpecialTxTable()
}

// After the scheduling is completed, get the result from the current snapshot
func (s *SnapshotEvidence) GetTxResultMap() map[string]*commonPb.Result {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetTxResultMap()
}

func (s *SnapshotEvidence) GetTxRWSetTable() []*commonPb.TxRWSet {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.GetTxRWSetTable()
}

func (s *SnapshotEvidence) GetKey(txExecSeq int, contractName string, key []byte) ([]byte, error) {
	if s.delegate == nil {
		return nil, errors.New("delegate is nil")
	}
	return s.delegate.GetKey(txExecSeq, contractName, key)
}

// After the read-write set is generated, add TxSimContext to the snapshot
// return if apply successfully or not, and current applied tx num
func (s *SnapshotEvidence) ApplyTxSimContext(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
	runVmSuccess bool, withSpecialTx bool) (bool, int) {
	if s.delegate == nil {
		return false, -1
	}
	return s.delegate.ApplyTxSimContext(txSimContext, specialTxType, runVmSuccess, withSpecialTx)
}

// check if snapshot is sealed
func (s *SnapshotEvidence) IsSealed() bool {
	if s.delegate == nil {
		return false
	}
	return s.delegate.IsSealed()

}

// get block height for current snapshot
func (s *SnapshotEvidence) GetBlockHeight() uint64 {
	if s.delegate == nil {
		return math.MaxUint64
	}
	return s.delegate.GetBlockHeight()
}

// get block height for current snapshot
func (s *SnapshotEvidence) GetBlockTimestamp() int64 {
	if s.delegate == nil {
		return math.MaxInt64
	}
	return s.delegate.GetBlockTimestamp()
}

// seal the snapshot
func (s *SnapshotEvidence) Seal() {
	if s.delegate == nil {
		return
	}
	s.delegate.Seal()
}

// According to the read-write table, the read-write dependency is checked from back to front to determine whether
// the transaction can be executed concurrently.
// From the process of building the read-write table, we have known that every transaction is based on a known
// world state, or cache state. As long as the world state or cache state that the tx depends on does not
// change during the execution, then the execution result of the transaction is determined.
// We need to ensure that when validating the DAG, there is no possibility that the execution of other
// transactions will affect the dependence of the current transaction
func (s *SnapshotEvidence) BuildDAG(isSql bool) *commonPb.DAG {
	if s.delegate == nil {
		return nil
	}
	if !s.IsSealed() {
		s.log.Warnf("you need to execute Seal before you can build DAG of snapshot with height %d", s.delegate.blockHeight)
	}
	s.delegate.lock.Lock()
	defer s.delegate.lock.Unlock()

	txCount := len(s.delegate.txTable)
	s.log.Debugf("start building DAG(all vertexes are nil) for block %d with %d txs", s.delegate.blockHeight, txCount)

	dag := &commonPb.DAG{}
	if txCount == 0 {
		return dag
	}

	dag.Vertexes = make([]*commonPb.DAG_Neighbor, txCount)

	if isSql {
		for i := 0; i < txCount; i++ {
			dag.Vertexes[i] = &commonPb.DAG_Neighbor{
				Neighbors: make([]uint32, 0, 1),
			}
			if i != 0 {
				dag.Vertexes[i].Neighbors = append(dag.Vertexes[i].Neighbors, uint32(i-1))
			}
		}
	} else {
		for i := 0; i < txCount; i++ {
			// build DAG based on directReach bitmap
			dag.Vertexes[i] = &commonPb.DAG_Neighbor{
				Neighbors: nil,
			}
		}
	}
	s.log.Debugf("build DAG for block %d finished", s.delegate.blockHeight)
	return dag
}

// Get Block Proposer for current snapshot
func (s *SnapshotEvidence) GetBlockProposer() *accesscontrol.Member {
	if s.delegate == nil {
		return nil
	}
	return s.delegate.blockProposer
}
