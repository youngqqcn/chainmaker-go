/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package solo

import (
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2/mock"
)

const (
	id     = "QmQZn3pZCcuEf34FSvucqkvVJEvfzpNjQTk17HS6CYMR35"
	org1Id = "wx-org1"
)

func TestNew(t *testing.T) {
	t.Log("TestNew")
	consensusSoloImplList := createConsensusSoloImpl(t)
	t.Log(consensusSoloImplList)
}

func TestStart(t *testing.T) {
	t.Log("TestStart")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		err := consensusSoloImpl.Start()
		if err != nil {
			t.Error(err)
			continue
		}
	}

	t.Log("TestStart pass")
}

// TODO stop func 没有完善
func TestStop(t *testing.T) {
	t.Log("TestStop")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImp := range consensusSoloImplList {
		err := consensusSoloImp.Stop()
		if err != nil {
			t.Error(err)
			continue
		}
	}

	t.Log("TestStop pass")
}

func TestOnMessage(t *testing.T) {
	t.Log("TestOnMessage")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImp := range consensusSoloImplList {
		consensusSoloImp.OnMessage(&msgbus.Message{
			Topic:   msgbus.BlockInfo,
			Payload: "666",
		})
	}

}

func TestOnQuit(t *testing.T) {
	t.Log("TestOnQuit")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		consensusSoloImpl.OnQuit()
	}

}

// TODO CanProposeBlock func 没有完善
func TestCanProposeBlock(t *testing.T) {
	t.Log("TestCanProposeBlock")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		res := consensusSoloImpl.CanProposeBlock()

		t.Log("TestCanProposeBlock res:", res)
	}
}

// TODO VerifyBlockSignatures func 没有完善
func TestVerifyBlockSignatures(t *testing.T) {
	t.Log("TestVerifyBlockSignatures")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		block := createNewBlock(1, 2)

		err := consensusSoloImpl.VerifyBlockSignatures(block)
		if err != nil {
			t.Error(err)
		}
	}

}

func TestProcProposerStatus(t *testing.T) {
	t.Log("TestProcProposerStatus")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		consensusSoloImpl.procProposerStatus()

	}

}

func TestHandleProposedBlock(t *testing.T) {

	t.Log("TestHandleProposedBlock")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for _, consensusSoloImpl := range consensusSoloImplList {
		block := createNewBlock(1, 2)

		consensusSoloImpl.verifyingBlock = block

		consensusSoloImpl.handleProposedBlock(&msgbus.Message{
			Topic: msgbus.BlockInfo,
		})
	}

}

func TestHandleVerifyResult(t *testing.T) {

	t.Log("TestHandleVerifyResult")

	consensusSoloImplList := createConsensusSoloImpl(t)

	for k, consensusSoloImpl := range consensusSoloImplList {

		if k < 2 {
			block := createNewBlock(1, 2)
			consensusSoloImpl.verifyingBlock = block
			consensusSoloImpl.handleVerifyResult(&msgbus.Message{
				Topic: msgbus.BlockInfo,
				Payload: &consensuspb.VerifyResult{
					VerifiedBlock: block,
				},
			})
		} else {
			consensusSoloImpl.verifyingBlock = nil
			consensusSoloImpl.handleVerifyResult(&msgbus.Message{
				Topic: msgbus.BlockInfo,
			})
		}

	}
}

func createNewBlock(height uint64, timeStamp int64) *commonPb.Block {
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			BlockHeight:    height,
			PreBlockHash:   nil,
			BlockHash:      nil,
			BlockVersion:   0,
			DagHash:        nil,
			RwSetRoot:      nil,
			BlockTimestamp: timeStamp,
			Proposer:       &accesscontrol.Member{MemberInfo: []byte{1, 2, 3}},
			ConsensusArgs:  nil,
			TxCount:        0,
			Signature:      nil,
			BlockType:      commonPb.BlockType_CONFIG_BLOCK,
		},
		Dag: &commonPb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}
	block.Header.PreBlockHash = nil
	return block
}

func createConsensusSoloImpl(t *testing.T) []*ConsensusSoloImpl {

	consensusSoloImplist := make([]*ConsensusSoloImpl, 0)

	for i := 0; i < 5; i++ {
		var (
			chainId   = "123456" + strconv.Itoa(i)
			uid       = "123" + strconv.Itoa(i)
			msgBusN   = msgbus.NewMessageBus()
			ctl       = gomock.NewController(t)
			chainConf = mock.NewMockChainConf(ctl)
			signer    = mock.NewMockSigningMember(ctl)
		)

		chainConf.EXPECT().ChainConfig().AnyTimes().Return(&configpb.ChainConfig{
			Consensus: &configpb.ConsensusConfig{
				Type: consensuspb.ConsensusType_TBFT,
				Nodes: []*configpb.OrgConfig{
					{
						OrgId:  org1Id,
						NodeId: []string{id},
					},
				},
			},
			Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
		})

		//signer.EXPECT().Sign(gomock.Any(), gomock.Any())
		//signer.Sign("SHA256", []byte("test tt"))

		consensusSoloImp, err := New(chainId, uid, signer, msgBusN, chainConf)

		if err != nil {
			t.Error(err)
			continue
		}

		consensusSoloImplist = append(consensusSoloImplist, consensusSoloImp)
	}

	return consensusSoloImplist
}
