/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consensus

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	dpos "chainmaker.org/chainmaker/consensus-dpos/v2"
	hotstuff "chainmaker.org/chainmaker/consensus-hotstuff/v2"
	raft "chainmaker.org/chainmaker/consensus-raft/v2"
	solo "chainmaker.org/chainmaker/consensus-solo/v2"
	tbft "chainmaker.org/chainmaker/consensus-tbft/v2"
	utils "chainmaker.org/chainmaker/consensus-utils/v2"
	"chainmaker.org/chainmaker/localconf/v2"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
)

const (
	id     = "QmQZn3pZCcuEf34FSvucqkvVJEvfzpNjQTk17HS6CYMR35"
	org1Id = "wx-org1"
)

type TestBlockchain struct {
	chainId       string
	msgBus        msgbus.MessageBus
	netService    protocol.NetService
	store         protocol.BlockchainStore
	coreEngine    protocol.CoreEngine
	identity      protocol.SigningMember
	ac            protocol.AccessControlProvider
	ledgerCache   protocol.LedgerCache
	proposalCache protocol.ProposalCache
	chainConf     protocol.ChainConf
}

func (bc *TestBlockchain) MockInit(ctrl *gomock.Controller, consensusType consensuspb.ConsensusType) {
	bc.identity = mock.NewMockSigningMember(ctrl)
	ledgerCache := mock.NewMockLedgerCache(ctrl)
	ledgerCache.EXPECT().CurrentHeight().AnyTimes().Return(uint64(1), nil)
	bc.ledgerCache = ledgerCache
	chainConf := mock.NewMockChainConf(ctrl)
	chainConf.EXPECT().ChainConfig().AnyTimes().Return(&configpb.ChainConfig{
		Consensus: &configpb.ConsensusConfig{
			Type: consensusType,
			Nodes: []*configpb.OrgConfig{
				{
					OrgId:  org1Id,
					NodeId: []string{id},
				},
			},
		},
	})
	bc.chainConf = chainConf
}

func TestNewConsensusEngine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prePath := localconf.ChainMakerConfig.GetStorePath()
	defer func() {
		localconf.ChainMakerConfig.StorageConfig["store_path"] = prePath
	}()
	localconf.ChainMakerConfig.StorageConfig["store_path"] = filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().Nanosecond()))

	tests := []struct {
		name    string
		csType  consensuspb.ConsensusType
		want    protocol.ConsensusEngine
		wantErr bool
	}{
		{"new TBFT consensus engine",
			consensuspb.ConsensusType_TBFT,
			&tbft.ConsensusTBFTImpl{},
			false,
		},
		{"new SOLO consensus engine",
			consensuspb.ConsensusType_SOLO,
			&solo.ConsensusSoloImpl{},
			false,
		},
		{"new RAFT consensus engine",
			consensuspb.ConsensusType_RAFT,
			&raft.ConsensusRaftImpl{},
			false,
		},
		{"new DPOS consensus engine",
			consensuspb.ConsensusType_DPOS,
			&dpos.DPoSImpl{},
			false,
		},
		{"new HOTSTUFF consensus engine",
			consensuspb.ConsensusType_HOTSTUFF,
			&hotstuff.ConsensusChainedBftImpl{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &TestBlockchain{}
			bc.MockInit(ctrl, tt.csType)
			provider := GetConsensusProvider(bc.chainConf.ChainConfig().Consensus.Type)
			config := &utils.ConsensusImplConfig{
				ChainId: bc.chainId,
				NodeId: id,
				Ac: bc.ac,
				Core: bc.coreEngine,
				ChainConf: bc.chainConf,
				Signer: bc.identity,
				Store: bc.store,
				LedgerCache: bc.ledgerCache,
				ProposalCache: bc.proposalCache,
				MsgBus: bc.msgBus,
			}
			got, err := provider(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConsensusEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewConsensusEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}
