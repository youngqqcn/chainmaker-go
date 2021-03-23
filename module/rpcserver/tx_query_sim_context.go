/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcserver

import (
	acPb "chainmaker.org/chainmaker-go/pb/protogo/accesscontrol"
	commonPb "chainmaker.org/chainmaker-go/pb/protogo/common"
	"chainmaker.org/chainmaker-go/protocol"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"strings"
)

// Storage interface for smart contracts, implement TxSimContext
type txQuerySimContextImpl struct {
	tx              *commonPb.Transaction
	txResult        *commonPb.Result
	txReadKeyMap    map[string]*commonPb.TxRead
	txWriteKeyMap   map[string]*commonPb.TxWrite
	blockchainStore protocol.BlockchainStore
	vmManager       protocol.VmManager
	gasUsed         uint64 // only for callContract
	currentDeep     int
	currentResult   []byte
	hisResult       []*callContractResult
}

type callContractResult struct {
	contractName string
	method       string
	param        map[string]string
	deep         int
	gasUsed      uint64
	result       []byte
}

// StateDB & ReadWriteSet
func (s *txQuerySimContextImpl) Get(contractName string, key []byte) ([]byte, error) {
	// Get from write set
	value, done := s.getFromWriteSet(contractName, key)
	if done {
		s.putIntoReadSet(contractName, key, value)
		return value, nil
	}

	// Get from read set
	value, done = s.getFromReadSet(contractName, key)
	if done {
		return value, nil
	}

	// Get from db
	value, err := s.blockchainStore.ReadObject(contractName, key)
	if err != nil {
		return nil, err
	}

	// if get from db success, put into read set
	s.putIntoReadSet(contractName, key, value)
	return value, nil
}

func (s *txQuerySimContextImpl) Put(contractName string, key []byte, value []byte) error {
	s.putIntoWriteSet(contractName, key, value)
	return nil
}

func (s *txQuerySimContextImpl) Del(contractName string, key []byte) error {
	s.putIntoWriteSet(contractName, key, nil)
	return nil
}

func (s *txQuerySimContextImpl) Select(contractName string, startKey []byte, limit []byte) (protocol.Iterator, error) {
	return s.blockchainStore.SelectObject(contractName, startKey, limit), nil
}

func (s *txQuerySimContextImpl) GetCreator(contractName string) *acPb.SerializedMember {
	if creatorByte, err := s.Get(contractName, []byte(protocol.ContractCreator)); err != nil {
		return nil
	} else {
		creator := &acPb.SerializedMember{}
		if err = proto.Unmarshal(creatorByte, creator); err != nil {
			return nil
		}
		return creator
	}
}

func (s *txQuerySimContextImpl) GetSender() *acPb.SerializedMember {
	return s.tx.Header.Sender
}

func (s *txQuerySimContextImpl) GetBlockHeight() int64 {
	if lastBlock, err := s.blockchainStore.GetLastBlock(); err != nil {
		return 0
	} else {
		return lastBlock.Header.BlockHeight
	}
}

func (s *txQuerySimContextImpl) putIntoReadSet(contractName string, key []byte, value []byte) {
	s.txReadKeyMap[constructKey(contractName, key)] = &commonPb.TxRead{
		Key:          key,
		Value:        value,
		ContractName: contractName,
		Version:      nil,
	}
}

func (s *txQuerySimContextImpl) putIntoWriteSet(contractName string, key []byte, value []byte) {
	s.txWriteKeyMap[constructKey(contractName, key)] = &commonPb.TxWrite{
		Key:          key,
		Value:        value,
		ContractName: contractName,
	}
}

func (s *txQuerySimContextImpl) getFromReadSet(contractName string, key []byte) ([]byte, bool) {
	if txRead, ok := s.txReadKeyMap[constructKey(contractName, key)]; ok {
		return txRead.Value, true
	}
	return nil, false
}

func (s *txQuerySimContextImpl) getFromWriteSet(contractName string, key []byte) ([]byte, bool) {
	if txWrite, ok := s.txWriteKeyMap[constructKey(contractName, key)]; ok {
		return txWrite.Value, true
	}
	return nil, false
}

func (s *txQuerySimContextImpl) GetTx() *commonPb.Transaction {
	return s.tx
}

func (s *txQuerySimContextImpl) GetBlockchainStore() protocol.BlockchainStore {
	return s.blockchainStore
}

// Get access control service
func (s *txQuerySimContextImpl) GetAccessControl() (protocol.AccessControlProvider, error) {
	if s.vmManager.GetAccessControl() == nil {
		return nil, errors.New("access control for tx sim context is nil")
	}
	return s.vmManager.GetAccessControl(), nil
}

func (s *txQuerySimContextImpl) GetContractTxIds() ([]string, error) {
	if txIdsBytes, err := s.Get(protocol.ContractTxIdsKey, nil); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get tx ids, error:%s", err.Error()))
	} else {
		txIds := strings.Split(string(txIdsBytes), ",")
		return txIds, nil
	}
}

// Get organization service
func (s *txQuerySimContextImpl) GetChainNodesInfoProvider() (protocol.ChainNodesInfoProvider, error) {
	if s.vmManager.GetChainNodesInfoProvider() == nil {
		return nil, errors.New("chainNodesInfoProvider for tx sim context is nil")
	}
	return s.vmManager.GetChainNodesInfoProvider(), nil
}

func (s *txQuerySimContextImpl) GetTxRWSet() *commonPb.TxRWSet {
	txRwSet := &commonPb.TxRWSet{
		TxId:     s.tx.Header.TxId,
		TxReads:  nil,
		TxWrites: nil,
	}
	for _, txRead := range s.txReadKeyMap {
		txRwSet.TxReads = append(txRwSet.TxReads, txRead)
	}
	for _, txWrite := range s.txWriteKeyMap {
		txRwSet.TxWrites = append(txRwSet.TxWrites, txWrite)
	}
	return txRwSet
}

func (s *txQuerySimContextImpl) GetTxExecSeq() int {
	return 0
}

func (s *txQuerySimContextImpl) SetTxExecSeq(int) {
	return
}

// Get the tx result
func (s *txQuerySimContextImpl) GetTxResult() *commonPb.Result {
	return s.txResult
}

// Set the tx result
func (s *txQuerySimContextImpl) SetTxResult(txResult *commonPb.Result) {
	s.txResult = txResult
}

func constructKey(contractName string, key []byte) string {
	return contractName + string(key)
}

func (s *txQuerySimContextImpl) CallContract(contractId *commonPb.ContractId, method string, byteCode []byte, parameter map[string]string, gasUsed uint64, refTxType commonPb.TxType) (*commonPb.ContractResult, commonPb.TxStatusCode) {
	s.gasUsed = gasUsed
	s.currentDeep = s.currentDeep + 1
	if s.currentDeep > protocol.CallContractDeep {
		contractResult := &commonPb.ContractResult{
			Code:    commonPb.ContractResultCode_FAIL,
			Result:  nil,
			Message: fmt.Sprintf("CallContract too deep %d", s.currentDeep),
		}
		return contractResult, commonPb.TxStatusCode_CONTRACT_TOO_DEEP_FAILED
	}
	if s.gasUsed > protocol.GasLimit {
		contractResult := &commonPb.ContractResult{
			Code:    commonPb.ContractResultCode_FAIL,
			Result:  nil,
			Message: fmt.Sprintf("There is not enough gas, gasUsed %d GasLimit %d ", gasUsed, int64(protocol.GasLimit)),
		}
		return contractResult, commonPb.TxStatusCode_CONTRACT_FAIL
	}
	r, code := s.vmManager.RunContract(contractId, method, byteCode, parameter, s, s.gasUsed, refTxType)

	result := callContractResult{
		deep:         s.currentDeep,
		gasUsed:      s.gasUsed,
		result:       r.Result,
		contractName: contractId.ContractName,
		method:       method,
		param:        parameter,
	}
	s.hisResult = append(s.hisResult, &result)
	s.currentResult = r.Result
	s.currentDeep = s.currentDeep - 1
	return r, code
}

func (s *txQuerySimContextImpl) GetCurrentResult() []byte {
	return s.currentResult
}

func (s *txQuerySimContextImpl) GetDepth() int {
	return s.currentDeep
}