/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	"fmt"
	"strings"
	"sync"

	"chainmaker.org/chainmaker/localconf/v2"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"go.uber.org/atomic"

	"chainmaker.org/chainmaker/common/v2/bitmap"
	"chainmaker.org/chainmaker/protocol/v2"
)

// The record value is written by the SEQ corresponding to TX
type sv struct {
	seq   int
	value []byte
}

type SnapshotImpl struct {
	lock            sync.RWMutex
	blockchainStore protocol.BlockchainStore
	log             protocol.Logger
	// If the snapshot has been sealed, the results of subsequent vm execution will not be added to the snapshot
	sealed *atomic.Bool

	chainId        string
	blockTimestamp int64
	blockProposer  *accesscontrol.Member
	blockHeight    uint64
	blockVersion   uint32
	preBlockHash   []byte

	preSnapshot protocol.Snapshot

	// applied data, please lock it before using
	txRWSetTable   []*commonPb.TxRWSet
	txTable        []*commonPb.Transaction
	specialTxTable []*commonPb.Transaction
	txResultMap    map[string]*commonPb.Result
	readTable      map[string]*sv
	writeTable     map[string]*sv

	txRoot    []byte
	dagHash   []byte
	rwSetHash []byte
}

func (s *SnapshotImpl) GetPreSnapshot() protocol.Snapshot {
	return s.preSnapshot
}

func (s *SnapshotImpl) SetPreSnapshot(snapshot protocol.Snapshot) {
	s.preSnapshot = snapshot
}

// GetBlockchainStore return the blockchainStore of the snapshot
func (s *SnapshotImpl) GetBlockchainStore() protocol.BlockchainStore {
	return s.blockchainStore
}

// GetSnapshotSize return the len of the txTable
func (s *SnapshotImpl) GetSnapshotSize() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.txTable)
}

// GetTxTable return the txTable of the snapshot
func (s *SnapshotImpl) GetTxTable() []*commonPb.Transaction {
	return s.txTable
}

// GetSpecialTxTable return the specialTxTable of the snapshot
func (s *SnapshotImpl) GetSpecialTxTable() []*commonPb.Transaction {
	return s.specialTxTable
}

// GetTxResultMap After the scheduling is completed, get the result from the current snapshot
func (s *SnapshotImpl) GetTxResultMap() map[string]*commonPb.Result {
	return s.txResultMap
}

// GetTxRWSetTable return the snapshot's txRWSetTable
func (s *SnapshotImpl) GetTxRWSetTable() []*commonPb.TxRWSet {
	if localconf.ChainMakerConfig.SchedulerConfig.RWSetLog {
		s.log.DebugDynamic(func() string {

			info := "rwset: "
			for i, txRWSet := range s.txRWSetTable {
				info += fmt.Sprintf("read set for tx id:[%s], count [%d]<", s.txTable[i].Payload.TxId, len(txRWSet.TxReads))
				//for _, txRead := range txRWSet.TxReads {
				//	if !strings.HasPrefix(string(txRead.Key), protocol.ContractByteCode) {
				//		info += fmt.Sprintf("[%v] -> [%v], contract name [%v], version [%v],",
				//		txRead.Key, txRead.Value, txRead.ContractName, txRead.Version)
				//	}
				//}
				info += "> "
				info += fmt.Sprintf("write set for tx id:[%s], count [%d]<", s.txTable[i].Payload.TxId, len(txRWSet.TxWrites))
				for _, txWrite := range txRWSet.TxWrites {
					info += fmt.Sprintf("[%v] -> [%v], contract name [%v], ", txWrite.Key, txWrite.Value, txWrite.ContractName)
				}
				info += ">"
			}
			return info
		})
		//log.Debugf(info)
	}

	//for _, txRWSet := range s.txRWSetTable {
	//	for _, txRead := range txRWSet.TxReads {
	//		if strings.HasPrefix(string(txRead.Key), protocol.ContractByteCode) ||
	//			strings.HasPrefix(string(txRead.Key), protocol.ContractCreator) ||
	//			txRead.ContractName == syscontract.SystemContract_CERT_MANAGE.String() {
	//			txRead.Value = nil
	//		}
	//	}
	//}
	return s.txRWSetTable
}

// GetKey from snapshot
func (s *SnapshotImpl) GetKey(txExecSeq int, contractName string, key []byte) ([]byte, error) {
	// get key before txExecSeq
	snapshotSize := s.GetSnapshotSize()

	s.lock.RLock()
	defer s.lock.RUnlock()
	if txExecSeq > snapshotSize || txExecSeq < 0 {
		txExecSeq = snapshotSize //nolint: ineffassign, staticcheck
	}
	finalKey := constructKey(contractName, key)
	if sv, ok := s.writeTable[finalKey]; ok {
		return sv.value, nil
	}

	if sv, ok := s.readTable[finalKey]; ok {
		return sv.value, nil
	}

	iter := s.preSnapshot
	for iter != nil {
		if value, err := iter.GetKey(-1, contractName, key); err == nil {
			return value, nil
		}
		iter = iter.GetPreSnapshot()
	}

	return s.blockchainStore.ReadObject(contractName, key)
}

// ApplyTxSimContext add TxSimContext to the snapshot, return current applied tx num whether success of not
func (s *SnapshotImpl) ApplyTxSimContext(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
	runVmSuccess bool, applySpecialTx bool) (bool, int) {
	tx := txSimContext.GetTx()
	s.log.Debugf("apply tx: %s, execOrderTxType:%d, runVmSuccess:%v, applySpecialTx:%v", tx.Payload.TxId,
		specialTxType, runVmSuccess, applySpecialTx)
	if !applySpecialTx && s.IsSealed() {
		return false, s.GetSnapshotSize()
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	// it is necessary to check sealed secondly
	if !applySpecialTx && s.IsSealed() {
		return false, len(s.txTable)
	}

	txExecSeq := txSimContext.GetTxExecSeq()
	var txRWSet *commonPb.TxRWSet
	var txResult *commonPb.Result

	if !applySpecialTx && specialTxType == protocol.ExecOrderTxTypeIterator {
		s.specialTxTable = append(s.specialTxTable, tx)
		return true, len(s.txTable) + len(s.specialTxTable)
	}

	// Only when the virtual machine is running normally can the read-write set be saved, or write fake conflicted key
	txRWSet = txSimContext.GetTxRWSet(runVmSuccess)
	txResult = txSimContext.GetTxResult()

	if specialTxType == protocol.ExecOrderTxTypeIterator || txExecSeq >= len(s.txTable) {
		s.apply(tx, txRWSet, txResult, runVmSuccess)
		return true, len(s.txTable)
	}

	// Check whether the dependent state has been modified during the running it
	for _, txRead := range txRWSet.TxReads {
		finalKey := constructKey(txRead.ContractName, txRead.Key)
		if sv, ok := s.writeTable[finalKey]; ok {
			if sv.seq >= txExecSeq {
				s.log.Debugf("Key Conflicted %+v-%+v, tx id:%s", sv.seq, txExecSeq, tx.Payload.TxId)
				return false, len(s.txTable)
			}
		}
	}

	s.apply(tx, txRWSet, txResult, runVmSuccess)
	return true, len(s.txTable)
}

// After the read-write set is generated, add TxSimContext to the snapshot
func (s *SnapshotImpl) apply(tx *commonPb.Transaction, txRWSet *commonPb.TxRWSet, txResult *commonPb.Result,
	runVmSuccess bool) {
	// Append to read table
	applySeq := len(s.txTable)
	// compatible with version lower than 2201
	if s.blockVersion < 2201 || runVmSuccess {
		for _, txRead := range txRWSet.TxReads {
			finalKey := constructKey(txRead.ContractName, txRead.Key)
			s.readTable[finalKey] = &sv{
				seq:   applySeq,
				value: txRead.Value,
			}
		}
	}

	// Append to write table
	for _, txWrite := range txRWSet.TxWrites {
		finalKey := constructKey(txWrite.ContractName, txWrite.Key)
		s.writeTable[finalKey] = &sv{
			seq:   applySeq,
			value: txWrite.Value,
		}
	}

	// Append to read-write-set table
	s.txRWSetTable = append(s.txRWSetTable, txRWSet)
	s.log.Debugf("apply tx: %s, rwset table size %d", tx.Payload.TxId, len(s.txRWSetTable))

	// Add to tx result map
	s.txResultMap[tx.Payload.TxId] = txResult

	// Add to transaction table
	s.txTable = append(s.txTable, tx)
}

// IsSealed check if snapshot is sealed
func (s *SnapshotImpl) IsSealed() bool {
	return s.sealed.Load()
}

// GetBlockHeight returns current block height
func (s *SnapshotImpl) GetBlockHeight() uint64 {
	return s.blockHeight
}

// GetBlockTimestamp returns current block timestamp
func (s *SnapshotImpl) GetBlockTimestamp() int64 {
	return s.blockTimestamp
}

// GetBlockProposer for current snapshot
func (s *SnapshotImpl) GetBlockProposer() *accesscontrol.Member {
	return s.blockProposer
}

// Seal the snapshot
func (s *SnapshotImpl) Seal() {
	s.sealed.Store(true)
}

// BuildDAG build the block dag according to the read-write table
func (s *SnapshotImpl) BuildDAG(isSql bool) *commonPb.DAG {
	s.lock.RLock()
	defer s.lock.RUnlock()

	txCount := uint32(len(s.txTable))
	s.log.Infof("start to build DAG for block %d with %d txs", s.blockHeight, txCount)
	dag := &commonPb.DAG{}
	if txCount == 0 {
		return dag
	}
	dag.Vertexes = make([]*commonPb.DAG_Neighbor, txCount)

	if isSql {
		for i := uint32(0); i < txCount; i++ {
			dag.Vertexes[i] = &commonPb.DAG_Neighbor{
				Neighbors: make([]uint32, 0, 1),
			}
			if i != 0 {
				dag.Vertexes[i].Neighbors = append(dag.Vertexes[i].Neighbors, uint32(i-1))
			}
		}
		return dag
	}
	// build all txs' readKeyDictionary, writeKeyDictionary, readPos(the pos in readKeyDictionary) and
	// writePos(the pos in writeKeyDictionary)
	readKeyDict, writeKeyDict, readPos, writePos := s.buildDictAndPos(txCount)
	reachMap := make([]*bitmap.Bitmap, txCount)
	// build vertexes
	for i := uint32(0); i < txCount; i++ {
		directReachMap := s.buildReachMap(i, readKeyDict, writeKeyDict, readPos, writePos, reachMap)
		dag.Vertexes[i] = &commonPb.DAG_Neighbor{
			Neighbors: make([]uint32, 0, 16),
		}
		for _, j := range directReachMap.Pos1() {
			dag.Vertexes[i].Neighbors = append(dag.Vertexes[i].Neighbors, uint32(j))
		}
	}
	s.log.Infof("build DAG for block %d finished", s.blockHeight)
	return dag
}

func (s *SnapshotImpl) buildDictAndPos(txCount uint32) (map[string][]uint32, map[string][]uint32,
	map[uint32]map[string]uint32, map[uint32]map[string]uint32) {
	readKeyDict := make(map[string][]uint32, 1024)
	writeKeyDict := make(map[string][]uint32, 1024)
	readPos := make(map[uint32]map[string]uint32)
	writePos := make(map[uint32]map[string]uint32)
	for i := uint32(0); i < txCount; i++ {
		readTableItemForI := s.txRWSetTable[i].TxReads
		writeTableItemForI := s.txRWSetTable[i].TxWrites
		readPos[i] = make(map[string]uint32)
		writePos[i] = make(map[string]uint32)
		// put all read key in to readKeyDict and set their pos into readPos and writePos
		for _, keyForI := range readTableItemForI {
			key := string(keyForI.Key)
			readPos[i][key] = uint32(len(readKeyDict[key]))
			writePos[i][key] = uint32(len(writeKeyDict[key]))
			readKeyDict[key] = append(readKeyDict[key], i)
		}
		// put all write key in to writeKeyDict and set their pos into readPos and writePos
		for _, keyForI := range writeTableItemForI {
			key := string(keyForI.Key)
			writePos[i][key] = uint32(len(writeKeyDict[key]))
			_, ok := readPos[i][key]
			if !ok {
				readPos[i][key] = uint32(len(readKeyDict[key]))
			}
			writeKeyDict[key] = append(writeKeyDict[key], i)
		}
	}
	return readKeyDict, writeKeyDict, readPos, writePos
}

func (s *SnapshotImpl) buildReachMap(i uint32, readKeyDict, writeKeyDict map[string][]uint32,
	readPos, writePos map[uint32]map[string]uint32, reachMap []*bitmap.Bitmap) *bitmap.Bitmap {
	readTableItemForI := s.txRWSetTable[i].TxReads
	writeTableItemForI := s.txRWSetTable[i].TxWrites
	allReachForI := &bitmap.Bitmap{}
	allReachForI.Set(int(i))
	directReachForI := &bitmap.Bitmap{}

	//ReadSet && WriteSet conflict
	for _, keyForI := range readTableItemForI {
		readKey := string(keyForI.Key)
		writeKeyTxs := writeKeyDict[readKey]
		if len(writeKeyTxs) == 0 {
			continue
		}
		// just check 1 write key before the tx because write keys all are conflict
		j := int(writePos[i][readKey]) - 1
		if j >= 0 && !allReachForI.Has(int(writeKeyTxs[j])) {
			directReachForI.Set(int(writeKeyTxs[j]))
			allReachForI.Or(reachMap[writeKeyTxs[j]])
		}
	}
	//WriteSet and (all ReadSet, WriteSet) conflict
	for _, keyForI := range writeTableItemForI {
		writeKey := string(keyForI.Key)
		readKeyTxs := readKeyDict[writeKey]
		if len(readKeyTxs) > 0 {
			// we should check all readKeyTxs because read keys has no conflict
			j := int(readPos[i][writeKey]) - 1
			for ; j >= 0; j-- {
				if !allReachForI.Has(int(readKeyTxs[j])) {
					directReachForI.Set(int(readKeyTxs[j]))
					allReachForI.Or(reachMap[readKeyTxs[j]])
				}
			}
		}
		writeKeyTxs := writeKeyDict[writeKey]
		if len(writeKeyTxs) == 0 {
			continue
		}
		// just check 1 write key before the tx because write keys all are conflict
		j := int(writePos[i][writeKey]) - 1
		if j >= 0 && !allReachForI.Has(int(writeKeyTxs[j])) {
			directReachForI.Set(int(writeKeyTxs[j]))
			allReachForI.Or(reachMap[writeKeyTxs[j]])
		}
	}
	reachMap[i] = allReachForI
	return directReachForI
}

func constructKey(contractName string, key []byte) string {
	var builder strings.Builder
	builder.WriteString(contractName)
	builder.Write(key)
	return builder.String()
}
