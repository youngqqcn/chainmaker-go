/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package snapshot

import (
	"strconv"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"chainmaker.org/chainmaker/vm/v2"
	"github.com/golang/mock/gomock"
	uatomic "go.uber.org/atomic"
)

var snapshotEvidence = &SnapshotEvidence{
	delegate: &SnapshotImpl{
		blockchainStore: nil,
		log:             &test.GoLogger{},
	},
	log: &test.GoLogger{},
}

func TestSetPreSnapshot(t *testing.T) {
	t.Log("TestSetPreSnapshot")

	snapshotList, _ := createNewBlockGroup()

	for _, snapshot := range snapshotList {
		snapshotEvidence.SetPreSnapshot(snapshot)
		t.Log("snapshotEvidence.SetPreSnapshot", snapshot)
	}
}

func TestGetPreSnapshot(t *testing.T) {
	t.Log("TestGetPreSnapshot")
	t.Logf("snapshotEvidence.GetPreSnapshot with no set %v", snapshotEvidence.GetPreSnapshot())

	snapshotList, _ := createNewBlockGroup()

	for _, snapshot := range snapshotList {
		snapshotEvidence.SetPreSnapshot(snapshot)
		t.Logf("SetPreSnapshot %v snapshotEvidence.GetPreSnapshot %v", snapshot, snapshotEvidence.GetPreSnapshot())
	}
}

func TestGetBlockchainStore(t *testing.T) {
	t.Log("TestGetBlockchainStore")

	snapshotEvidenceList := createSnapshotEvidenceList(t)

	for _, v := range snapshotEvidenceList {
		t.Log(v.GetBlockchainStore())
	}
}

func TestGetSnapshotSize(t *testing.T) {
	t.Log("TestGetSnapshotSize")

	snapshotEvidenceList := createSnapshotEvidenceList(t)

	for _, v := range snapshotEvidenceList {
		t.Logf("snapshot %v, SnapshotSize %v", v, v.GetSnapshotSize())
	}
}

func TestGetTxTable(t *testing.T) {
	t.Log("TestGetTxTable")

	snapshotEvidenceList := createSnapshotEvidenceList(t)

	for _, v := range snapshotEvidenceList {
		t.Logf("snapshot %v, GetTxTable %v", v, v.GetTxTable())
	}
}

func TestGetTxResultMap(t *testing.T) {

	t.Log("TestGetTxResultMap")

	snapshotEvidenceList := createSnapshotEvidenceList(t)

	for _, v := range snapshotEvidenceList {
		t.Logf("snapshot %v, GetTxResultMap %v", v, v.GetTxResultMap())
	}
}

// TODO GetTxRWSetTable() 代码不完善。完善后再补充
func TestGetTxRWSetTable(t *testing.T) {
	t.Log("TestGetTxRWSetTable")
}

// TODO 没有输出key的信息
//func TestGetKey(t *testing.T) {
//	t.Log("TestGetKey")
//
//	snapshotEvidenceList := createSnapshotEvidenceList(t)
//
//	for _, v := range snapshotEvidenceList {
//
//		keyByte, err := v.GetKey(2, "TestGetKey", []byte("12345"))
//
//		if err != nil {
//			t.Errorf("v.GetKey err:%v", err)
//		}
//		t.Logf("snapshot %v, GetSnapshotSize %v, GetKey %v", v, v.GetSnapshotSize(), keyByte)
//	}
//}

func TestApplyTxSimContext(t *testing.T) {
	t.Log("TestApplyTxSimContext")

	// block构造txs
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			BlockHeight:    2,
			PreBlockHash:   nil,
			BlockHash:      nil,
			BlockVersion:   0,
			DagHash:        nil,
			RwSetRoot:      nil,
			BlockTimestamp: 1,
			Proposer:       &accesscontrol.Member{MemberInfo: []byte{1, 2, 3}},
			ConsensusArgs:  nil,
			TxCount:        0,
			Signature:      nil,
		},
		Dag: &commonPb.DAG{
			Vertexes: nil,
		},
		Txs: []*common.Transaction{{
			Payload: &common.Payload{ChainId: "12345"},
		}, {
			Payload: &common.Payload{ChainId: "678910"},
		}, {
			Payload: &common.Payload{ChainId: "1112131415"},
		}, {
			Payload: &common.Payload{ChainId: "1617181920"},
		}},
	}
	block.Header.BlockVersion = 12

	snapshot := createSnapshotEvidenceList(t)[0]
	vmManager := &mock.MockVmManager{}

	txSimContext := vm.NewTxSimContext(vmManager, snapshot,
		&common.Transaction{Payload: &common.Payload{ChainId: "12345"}},
		block.Header.BlockVersion, &test.GoLogger{})

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_ARCHIVED_BLOCK})
	res, tableLen := snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_ARCHIVED_BLOCK})
	res, tableLen = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, false, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_CONTRACT_FAIL})
	res, tableLen = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_CONTRACT_FAIL})
	res, tableLen = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, false, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_CONTRACT_REVOKE_FAILED})
	res, tableLen = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

	txSimContext.SetTxResult(&common.Result{Code: common.TxStatusCode_CONTRACT_REVOKE_FAILED})
	res, tableLen = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, false, false)
	t.Logf("snapshot.ApplyTxSimContext res:%v,tableLen:%v", res, tableLen)

}

func createSnapshotEvidenceList(t *testing.T) []*SnapshotEvidence {

	snapshotEvidenceList := make([]*SnapshotEvidence, 0)

	txTable := make([]*common.Transaction, 0)
	for i := 0; i < 5; i++ {
		ctl := gomock.NewController(t)
		blockchainStore := mock.NewMockBlockchainStore(ctl)

		// readObject
		//blockchainStore.EXPECT().ReadObject("TestGetKey", []byte("12345"))
		txTable = append(txTable, &common.Transaction{
			Payload:   &common.Payload{},
			Sender:    common.Transaction{}.Sender,
			Endorsers: []*common.EndorsementEntry{},
			Result:    &common.Result{},
		})

		code := common.TxStatusCode_ARCHIVED_TX
		switch i {
		case 1:
			code = common.TxStatusCode_CONTRACT_FAIL
		case 2:
			code = common.TxStatusCode_ARCHIVED_BLOCK
		case 3:
			code = common.TxStatusCode_CONTRACT_REVOKE_FAILED
		case 4:
			code = common.TxStatusCode_CONTRACT_TOO_DEEP_FAILED
		}

		snapshotEvidence := &SnapshotEvidence{
			delegate: &SnapshotImpl{
				blockchainStore: blockchainStore,
				txTable:         txTable, // 0, 1, 2, 3, 4
				txResultMap: map[string]*common.Result{
					strconv.Itoa(i): {
						Code: code,
					},
				},
				log:    &test.GoLogger{},
				sealed: uatomic.NewBool(false),
			},
			log: &test.GoLogger{},
		}

		snapshotEvidenceList = append(snapshotEvidenceList, snapshotEvidence)
	}

	return snapshotEvidenceList
}
