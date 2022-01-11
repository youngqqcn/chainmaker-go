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
	maxbft "chainmaker.org/chainmaker/consensus-maxbft/v2"
	raft "chainmaker.org/chainmaker/consensus-raft/v2"
	solo "chainmaker.org/chainmaker/consensus-solo/v2"
	tbft "chainmaker.org/chainmaker/consensus-tbft/v2"
	utils "chainmaker.org/chainmaker/consensus-utils/v2"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	maxbftpb "chainmaker.org/chainmaker/pb-go/v2/consensus/maxbft"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/gogo/protobuf/proto"

	"github.com/golang/mock/gomock"
)

const (
	id     = "QmQZn3pZCcuEf34FSvucqkvVJEvfzpNjQTk17HS6CYMR35"
	org1Id = "wx-org1"
)

type TestBlockchain struct {
	chainId       string
	msgBus        msgbus.MessageBus
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
	qc := &maxbftpb.QuorumCert{
		BlockId: []byte("32c8b26"),
		Level:   0,
		Height:  0,
	}
	bytesQc, _ := proto.Marshal(qc)
	ledgerCache.EXPECT().GetLastCommittedBlock().AnyTimes().Return(&common.Block{
		Header: &common.BlockHeader{
			BlockHeight: 1,
			BlockHash:   []byte("77150e"),
		},
		AdditionalData: &common.AdditionalData{
			ExtraData: map[string][]byte{"QC": bytesQc},
		},
	})
	bc.ledgerCache = ledgerCache
	chainConf := mock.NewMockChainConf(ctrl)
	chainConf.EXPECT().ChainConfig().AnyTimes().Return(&configpb.ChainConfig{
		ChainId: id,
		Consensus: &configpb.ConsensusConfig{
			Type: consensusType,
			Nodes: []*configpb.OrgConfig{
				{
					OrgId:  org1Id,
					NodeId: []string{id},
				},
			},
		},
		Contract: &configpb.ContractConfig{
			EnableSqlSupport: false,
		},
	})
	bc.chainConf = chainConf
	coreEngine := mock.NewMockCoreEngine(ctrl)
	coreEngine.EXPECT().GetBlockVerifier().AnyTimes().Return(nil)
	coreEngine.EXPECT().GetBlockCommitter().AnyTimes().Return(nil)
	coreEngine.EXPECT().GetMaxbftHelper().AnyTimes().Return(nil)
	bc.coreEngine = coreEngine
	store := mock.NewMockBlockchainStore(ctrl)
	store.EXPECT().ReadObject("GOVERNANCE", []byte("GOVERNANCE")).AnyTimes().Return(
		[]byte{24, 4, 56, 4, 64, 3, 104, 4, 112, 1, 130, 1, 48, 10, 46, 81, 109, 82, 109, 54, 111,
			68, 121, 99, 111, 78, 71, 52, 56, 118, 69, 88, 104, 67, 118, 56, 99, 53, 86, 71, 110, 65,
			113, 75, 114, 107, 68, 98, 72, 102, 114, 72, 88, 118, 55, 104, 109, 87, 69, 115, 50, 130,
			1, 50, 10, 46, 81, 109, 82, 113, 74, 118, 56, 78, 70, 55, 65, 122, 52, 83, 119, 85, 71, 54,
			85, 113, 50, 71, 107, 67, 88, 50, 56, 102, 109, 113, 81, 56, 75, 99, 69, 74, 105, 49, 114, 51,
			56, 97, 112, 70, 52, 87, 16, 1, 130, 1, 50, 10, 46, 81, 109, 97, 78, 98, 121, 84, 104, 51, 110,
			57, 49, 99, 101, 50, 121, 121, 81, 107, 65, 111, 97, 105, 74, 121, 111, 70, 54, 111, 54, 51, 122,
			55, 55, 69, 75, 76, 103, 105, 109, 84, 56, 119, 78, 75, 103, 16, 2, 130, 1, 50, 10, 46, 81, 109,
			97, 86, 97, 55, 74, 75, 117, 101, 54, 74, 75, 82, 74, 90, 113, 121, 106, 77, 97, 77, 57, 104, 104,
			119, 122, 111, 56, 57, 105, 56, 83, 55, 107, 104, 85, 99, 111, 122, 83, 78, 76, 118, 65, 115, 16, 3}, nil)
	bc.store = store
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
		{"new MAXBFT consensus engine",
			consensuspb.ConsensusType_MAXBFT,
			&maxbft.ConsensusMaxBftImpl{},
			false,
		},
	}
	registerConsensuses()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &TestBlockchain{}
			bc.MockInit(ctrl, tt.csType)
			provider := GetConsensusProvider(bc.chainConf.ChainConfig().Consensus.Type)
			config := &utils.ConsensusImplConfig{
				ChainId:       bc.chainId,
				NodeId:        id,
				Ac:            bc.ac,
				Core:          bc.coreEngine,
				ChainConf:     bc.chainConf,
				Signer:        bc.identity,
				Store:         bc.store,
				LedgerCache:   bc.ledgerCache,
				ProposalCache: bc.proposalCache,
				MsgBus:        bc.msgBus,
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

func registerConsensuses() {
	// consensus
	RegisterConsensusProvider(
		consensuspb.ConsensusType_SOLO,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return solo.New(config)
		},
	)

	RegisterConsensusProvider(
		consensuspb.ConsensusType_DPOS,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			tbftEngine, err := tbft.New(config) // DPoS based in TBFT
			if err != nil {
				return nil, err
			}
			dposEngine := dpos.NewDPoSImpl(config, tbftEngine)
			return dposEngine, nil
		},
	)

	RegisterConsensusProvider(
		consensuspb.ConsensusType_RAFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return raft.New(config)
		},
	)

	RegisterConsensusProvider(
		consensuspb.ConsensusType_TBFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return tbft.New(config)
		},
	)

	RegisterConsensusProvider(
		consensuspb.ConsensusType_MAXBFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return maxbft.New(config)
		},
	)
}
