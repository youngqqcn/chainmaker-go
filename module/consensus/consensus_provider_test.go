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
	hotstuff "chainmaker.org/chainmaker/consensus-chainedbft/v2"
	dpos "chainmaker.org/chainmaker/consensus-dpos/v2"
	raft "chainmaker.org/chainmaker/consensus-raft/v2"
	solo "chainmaker.org/chainmaker/consensus-solo/v2"
	tbft "chainmaker.org/chainmaker/consensus-tbft/v2"
	utils "chainmaker.org/chainmaker/consensus-utils/v2"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	chainedbftpb "chainmaker.org/chainmaker/pb-go/v2/consensus/chainedbft"
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
	qc := &chainedbftpb.QuorumCert{
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
	coreEngine.EXPECT().GetHotStuffHelper().AnyTimes().Return(nil)
	bc.coreEngine = coreEngine
	store := mock.NewMockBlockchainStore(ctrl)
	store.EXPECT().ReadObject("GOVERNANCE", []byte{71, 79, 86, 69, 82, 78, 65, 78, 67, 69}).AnyTimes().Return(
		[]byte([]byte{71, 79, 86, 69, 82, 78, 65, 78, 67, 69}), nil)
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
		{"new HOTSTUFF consensus engine",
			consensuspb.ConsensusType_HOTSTUFF,
			&hotstuff.ConsensusChainedBftImpl{},
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
		consensuspb.ConsensusType_HOTSTUFF,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return hotstuff.New(config)
		},
	)
}
