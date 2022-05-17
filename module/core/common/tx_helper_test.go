/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"reflect"
	"testing"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/common"

	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/golang/mock/gomock"
)

//
//import (
//	"chainmaker.org/chainmaker/logger/v2"
//	"chainmaker.org/chainmaker/protocol/v2/mock"
//	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
//	"chainmaker.org/chainmaker/pb-go/v2/config"
//	"encoding/hex"
//	"fmt"
//	"github.com/golang/mock/gomock"
//	"testing"
//)
//
//func TestValidateTx(t *testing.T) {
//	verifyTx, block := txPrepare(t)
//	hashes, _, _, _ := verifyTx.verifierTxs(block)
//
//	for _, hash := range hashes {
//		fmt.Println("test hash: ", hex.EncodeToString(hash))
//	}
//}
//
//func txPrepare(t *testing.T) (*VerifierTx, *commonpb.Block) {
//	block := newBlock()
//	contractId := &commonpb.Contract{
//		ContractName:    "ContractName",
//		ContractVersion: "1",
//		RuntimeType:     commonpb.RuntimeType_WASMER,
//	}
//
//	parameters := make(map[string]string, 8)
//	tx0 := newTx("a0000000000000000000000000000000", contractId, parameters)
//	txs := make([]*commonpb.Transaction, 0)
//	txs = append(txs, tx0)
//	block.Txs = txs
//
//	var txRWSetMap = make(map[string]*commonpb.TxRWSet, 3)
//	txRWSetMap[tx0.Payload.TxId] = &commonpb.TxRWSet{
//		TxId: tx0.Payload.TxId,
//		TxReads: []*commonpb.TxRead{{
//			ContractName: contractId.Name,
//			Key:          []byte("K1"),
//			Value:        []byte("V"),
//		}},
//		TxWrites: []*commonpb.TxWrite{{
//			ContractName: contractId.Name,
//			Key:          []byte("K2"),
//			Value:        []byte("V"),
//		}},
//	}
//
//	rwHash, _ := hex.DecodeString("d02f421ed76e0e26e9def824a8b84c7c223d484762d6d060a8b71e1649d1abbf")
//	result := &commonpb.Result{
//		Code: commonpb.TxStatusCode_SUCCESS,
//		ContractResult: &commonpb.ContractResult{
//			Code:    0,
//			Result:  nil,
//			Message: "",
//			GasUsed: 0,
//		},
//		RwSetHash: rwHash,
//	}
//	tx0.Result = result
//	txResultMap := make(map[string]*commonpb.Result, 1)
//	txResultMap[tx0.Payload.TxId] = result
//
//	log := logger.GetLoggerByChain(logger.MODULE_CORE, "chain1")
//
//	ctl := gomock.NewController(t)
//	store := mock.NewMockBlockchainStore(ctl)
//	txPool := mock.NewMockTxPool(ctl)
//	ac := mock.NewMockAccessControlProvider(ctl)
//	chainConf := mock.NewMockChainConf(ctl)
//
//	store.EXPECT().TxExists(tx0).AnyTimes().Return(false, nil)
//
//	txsMap := make(map[string]*commonpb.Transaction)
//
//	txsMap[tx0.Payload.TxId] = tx0
//
//	txPool.EXPECT().GetTxsByTxIds([]string{tx0.Payload.TxId}).Return(txsMap, nil)
//	config := &config.ChainConfig{
//		ChainId: "chain1",
//		Crypto: &config.CryptoConfig{
//			Hash: "SHA256",
//		},
//	}
//	chainConf.EXPECT().ChainConfig().AnyTimes().Return(config)
//
//	principal := mock.NewMockPrincipal(ctl)
//	ac.EXPECT().LookUpResourceNameByTxType(tx0.Header.TxType).AnyTimes().Return("123", nil)
//	ac.EXPECT().CreatePrincipal("123", nil, nil).AnyTimes().Return(principal, nil)
//	ac.EXPECT().VerifyPrincipal(principal).AnyTimes().Return(true, nil)
//	verifyTxConf := &VerifierTxConfig{
//		Block:       block,
//		TxRWSetMap:  txRWSetMap,
//		TxResultMap: txResultMap,
//		Store:       store,
//		TxPool:      txPool,
//		Ac:          ac,
//		ChainConf:   chainConf,
//		Log:         log,
//	}
//	return NewVerifierTx(verifyTxConf), block
//}

func TestValidateTxRules(t *testing.T) {
	var txs []*common.Transaction
	for i := 0; i < 100; i++ {
		txs = append(txs, &common.Transaction{
			Payload: &common.Payload{
				TxId: utils.GetTimestampTxId(),
			},
		})
	}
	type args struct {
		filter protocol.TxFilter
		txs    []*common.Transaction
	}
	tests := []struct {
		name          string
		args          args
		wantRemoveTxs []*common.Transaction
		wantRemainTxs []*common.Transaction
	}{
		{
			name: "正常流",
			args: args{
				filter: func() protocol.TxFilter {
					filter := mock.NewMockTxFilter(gomock.NewController(t))
					filter.EXPECT().ValidateRule(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(txId string, ruleType ...common.RuleType) error {
						if []byte(txId)[63]%2 == 1 {
							return nil
						}
						return bn.ErrKeyTimeIsNotInTheFilterRange
					})
					return filter
				}(),
				txs: txs,
			},
			wantRemoveTxs: func() []*common.Transaction {
				var arr []*common.Transaction
				for _, tx := range txs {
					if []byte(tx.Payload.TxId)[63]%2 == 0 {
						arr = append(arr, tx)
					}
				}
				return arr
			}(),
			wantRemainTxs: func() []*common.Transaction {
				var arr []*common.Transaction
				for _, tx := range txs {
					if []byte(tx.Payload.TxId)[63]%2 == 1 {
						arr = append(arr, tx)
					}
				}
				return arr
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemoveTxs, gotRemainTxs := ValidateTxRules(tt.args.filter, tt.args.txs)
			if !reflect.DeepEqual(gotRemoveTxs, tt.wantRemoveTxs) {
				t.Errorf("ValidateTxRules() gotRemoveTxs = %v, want %v", gotRemoveTxs, tt.wantRemoveTxs)
			}
			if !reflect.DeepEqual(gotRemainTxs, tt.wantRemainTxs) {
				t.Errorf("ValidateTxRules() gotRemainTxs = %v, want %v", gotRemainTxs, tt.wantRemainTxs)
			}
		})
	}
}
