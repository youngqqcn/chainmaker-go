/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"
	"testing"
	"time"

	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//  statistic the time consuming of finalizeBlock between sync and async
// logLevel: Debug TxNum: 1000000; async:3037 ; sync: 4264
// logLevel: Info  TxNum: 1000000; async:224 ; sync: 251
func TestFinalizeBlock_Async(t *testing.T) {

	log := &test.HoleLogger{}
	block := createBlock(10)
	txs := make([]*commonPb.Transaction, 0)
	txRWSetMap := make(map[string]*commonPb.TxRWSet)
	for i := 0; i < 100000; i++ {
		txId := "0x123456789" + fmt.Sprint(i)
		tx := createNewTestTx(txId)
		txs = append(txs, tx)
		txRWSetMap[txId] = &commonPb.TxRWSet{
			TxId:    txId,
			TxReads: nil,
			TxWrites: []*commonPb.TxWrite{{
				Key:          []byte(fmt.Sprintf("key%d", i)),
				Value:        []byte(fmt.Sprintf("value[%d]", i)),
				ContractName: "TestContract",
			}},
		}
	}
	block.Txs = txs
	var err error

	asyncTimeStart := CurrentTimeMillisSeconds()
	err = FinalizeBlockSequence(block, txRWSetMap, nil, "SM3", log)
	t.Logf("sync mode cost:[%d]", CurrentTimeMillisSeconds()-asyncTimeStart)
	t.Logf("%x,%x,%x", block.Header.RwSetRoot, block.Header.TxRoot, block.Header.DagHash)
	rwSetRoot := block.Header.RwSetRoot
	//blockHash := block.Header.BlockHash
	dagHash := block.Header.DagHash
	txRoot := block.Header.TxRoot
	asyncTimeStart = CurrentTimeMillisSeconds()
	block.Header.RwSetRoot = nil
	block.Header.BlockHash = nil
	block.Header.DagHash = nil
	block.Header.TxRoot = nil
	err = FinalizeBlock(block, txRWSetMap, nil, "SM3", log)
	asyncTimeEnd := CurrentTimeMillisSeconds()
	require.Equal(t, nil, err)
	t.Logf("concurrent mode cost:[%d]", asyncTimeEnd-asyncTimeStart)
	assert.EqualValues(t, rwSetRoot, block.Header.RwSetRoot, "RwSetRoot")
	//assert.EqualValues(t, blockHash, block.Header.BlockHash, "BlockHash")
	assert.EqualValues(t, dagHash, block.Header.DagHash, "DagHash")
	assert.EqualValues(t, txRoot, block.Header.TxRoot, "TxRoot")

	////
	//syncTimeStart := CurrentTimeMillisSeconds()
	//err = FinalizeBlock(block, txRWSetMap, nil, "SHA256", log)
	//syncTimeEnd := CurrentTimeMillisSeconds()
	//require.Equal(t, nil, err)
	//t.Log(fmt.Sprintf("sync mode cost:[%d]", syncTimeEnd-syncTimeStart))
	////
	//require.Equal(t, rwSetRoot, block.Header.RwSetRoot)
	//require.Equal(t, blockHash, block.Header.BlockHash)
	//require.Equal(t, dagHash, block.Header.DagHash)
	//
	//log.Infof(fmt.Sprintf("async mode cost:[%d], sync mode cost:[%d]", asyncTimeEnd-asyncTimeStart, syncTimeEnd-syncTimeStart))

}

func createBlock(height uint64) *commonPb.Block {
	var hash = []byte("0123456789")
	var version = uint32(1)
	var block = &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:        "Chain1",
			BlockHeight:    height,
			PreBlockHash:   hash,
			BlockHash:      hash,
			PreConfHeight:  0,
			BlockVersion:   version,
			DagHash:        hash,
			RwSetRoot:      hash,
			TxRoot:         hash,
			BlockTimestamp: 0,
			Proposer:       &accesscontrol.Member{MemberInfo: hash},
			ConsensusArgs:  nil,
			TxCount:        1,
			Signature:      []byte(""),
		},
		Dag: &commonPb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}

	return block
}

func createNewTestTx(txID string) *commonPb.Transaction {
	//var hash = []byte("0123456789")
	return &commonPb.Transaction{
		Payload: &commonPb.Payload{
			ChainId:        "Chain1",
			TxType:         0,
			TxId:           txID,
			Timestamp:      CurrentTimeMillisSeconds(),
			ExpirationTime: 0,
		},
		//RequestPayload:   hash,
		//RequestSignature: hash,
		Result: &commonPb.Result{
			Code:           commonPb.TxStatusCode_SUCCESS,
			ContractResult: nil,
			RwSetHash:      nil,
		},
	}
}

func CurrentTimeMillisSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

// the sync way fo finalize block
func FinalizeBlockSync(
	block *commonPb.Block,
	txRWSetMap map[string]*commonPb.TxRWSet,
	aclFailTxs []*commonPb.Transaction,
	hashType string,
	logger protocol.Logger) error {

	if aclFailTxs != nil && len(aclFailTxs) > 0 { //nolint: gosimple
		// append acl check failed txs to the end of block.Txs
		block.Txs = append(block.Txs, aclFailTxs...)
	}

	// TxCount contains acl verify failed txs and invoked contract txs
	txCount := len(block.Txs)
	block.Header.TxCount = uint32(txCount)

	// TxRoot/RwSetRoot
	var err error
	txHashes := make([][]byte, txCount)
	for i, tx := range block.Txs {
		// finalize tx, put rwsethash into tx.Result
		rwSet := txRWSetMap[tx.Payload.TxId]
		if rwSet == nil {
			rwSet = &commonPb.TxRWSet{
				TxId:     tx.Payload.TxId,
				TxReads:  nil,
				TxWrites: nil,
			}
		}
		var rwSetHash []byte
		rwSetHash, err = utils.CalcRWSetHash(hashType, rwSet)
		logger.DebugDynamic(func() string {
			return fmt.Sprintf("CalcRWSetHash rwset: %+v ,hash: %x", rwSet, rwSetHash)
		})
		if err != nil {
			return err
		}
		if tx.Result == nil {
			// in case tx.Result is nil, avoid panic
			e := fmt.Errorf("tx(%s) result == nil", tx.Payload.TxId)
			logger.Error(e.Error())
			return e
		}
		tx.Result.RwSetHash = rwSetHash
		// calculate complete tx hash, include tx.Header, tx.Payload, tx.Result
		var txHash []byte
		txHash, err = utils.CalcTxHash(hashType, tx)
		if err != nil {
			return err
		}
		txHashes[i] = txHash
	}

	block.Header.TxRoot, err = hash.GetMerkleRoot(hashType, txHashes)
	if err != nil {
		logger.Warnf("get tx merkle root error %s", err)
		return err
	}
	block.Header.RwSetRoot, err = utils.CalcRWSetRoot(hashType, block.Txs)
	if err != nil {
		logger.Warnf("get rwset merkle root error %s", err)
		return err
	}

	// DagDigest
	dagHash, err := utils.CalcDagHash(hashType, block.Dag)
	if err != nil {
		logger.Warnf("get dag hash error %s", err)
		return err
	}
	block.Header.DagHash = dagHash

	return nil
}

//FinalizeBlockSequence 串行化的方式计算各个Hash，主要为了验证并行模式下结果的正确性和性能
func FinalizeBlockSequence(
	block *commonPb.Block,
	txRWSetMap map[string]*commonPb.TxRWSet,
	aclFailTxs []*commonPb.Transaction,
	hashType string,
	logger protocol.Logger) error {

	if aclFailTxs != nil && len(aclFailTxs) > 0 { //nolint: gosimple
		// append acl check failed txs to the end of block.Txs
		block.Txs = append(block.Txs, aclFailTxs...)
	}

	// TxCount contains acl verify failed txs and invoked contract txs
	txCount := len(block.Txs)
	block.Header.TxCount = uint32(txCount)

	// TxRoot/RwSetRoot
	var err error
	txHashes := make([][]byte, txCount)
	for i, tx := range block.Txs {
		// finalize tx, put rwsethash into tx.Result
		rwSet := txRWSetMap[tx.Payload.TxId]
		if rwSet == nil {
			rwSet = &commonPb.TxRWSet{
				TxId:     tx.Payload.TxId,
				TxReads:  nil,
				TxWrites: nil,
			}
		}

		var rwSetHash []byte
		rwSetHash, err = utils.CalcRWSetHash(hashType, rwSet)
		logger.DebugDynamic(func() string {
			str := fmt.Sprintf("CalcRWSetHash rwset: %+v ,hash: %x", rwSet, rwSetHash)
			if len(str) > 1024 {
				str = str[:1024] + " ......"
			}
			return str
		})
		if err != nil {
			return err
		}
		if tx.Result == nil {
			// in case tx.Result is nil, avoid panic
			e := fmt.Errorf("tx(%s) result == nil", tx.Payload.TxId)
			logger.Error(e.Error())
			return e
		}
		tx.Result.RwSetHash = rwSetHash
		// calculate complete tx hash, include tx.Header, tx.Payload, tx.Result
		var txHash []byte
		txHash, err = utils.CalcTxHash(hashType, tx)
		if err != nil {
			return err
		}
		txHashes[i] = txHash
	}

	block.Header.TxRoot, err = hash.GetMerkleRoot(hashType, txHashes)
	if err != nil {
		logger.Warnf("get tx merkle root error %s", err)
		return err
	}
	logger.DebugDynamic(func() string {
		return fmt.Sprintf("GetMerkleRoot(%s) get %x", hashType, block.Header.TxRoot)
	})
	block.Header.RwSetRoot, err = utils.CalcRWSetRoot(hashType, block.Txs)
	if err != nil {
		logger.Warnf("get rwset merkle root error %s", err)
		return err
	}

	// DagDigest
	var dagHash []byte
	dagHash, err = utils.CalcDagHash(hashType, block.Dag)
	if err != nil {
		logger.Warnf("get dag hash error %s", err)
		return err
	}
	block.Header.DagHash = dagHash

	return nil
}
