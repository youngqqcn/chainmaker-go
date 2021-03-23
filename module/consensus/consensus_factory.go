/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consensus

import (
	"fmt"

	commonpb "chainmaker.org/chainmaker-go/pb/protogo/common"
	consensuspb "chainmaker.org/chainmaker-go/pb/protogo/consensus"

	"chainmaker.org/chainmaker-go/common/msgbus"
	"chainmaker.org/chainmaker-go/consensus/raft"
	"chainmaker.org/chainmaker-go/consensus/solo"
	"chainmaker.org/chainmaker-go/consensus/tbft"
	"chainmaker.org/chainmaker-go/protocol"
)

type Factory struct {
}

// NewConsensusEngine new the consensus engine.
// consensusType specfies the consensus engine type.
// msgBus is used for send and receive messages.
func (f Factory) NewConsensusEngine(
	consensusType consensuspb.ConsensusType,
	chainID string,
	id string,
	nodeList []string,
	signer protocol.SigningMember,
	ac protocol.AccessControlProvider,
	dbHandle protocol.DBHandle,
	ledgerCache protocol.LedgerCache,
	proposalCache protocol.ProposalCache,
	blockVerifier protocol.BlockVerifier,
	blockCommitter protocol.BlockCommitter,
	netService protocol.NetService,
	msgBus msgbus.MessageBus,
	chainConf protocol.ChainConf,
	store protocol.BlockchainStore) (protocol.ConsensusEngine, error) {
	switch consensusType {
	case consensuspb.ConsensusType_TBFT:
		config := tbft.ConsensusTBFTImplConfig{
			ChainID:     chainID,
			Id:          id,
			Signer:      signer,
			Ac:          ac,
			DbHandle:    dbHandle,
			LedgerCache: ledgerCache,
			ChainConf:   chainConf,
			NetService:  netService,
			MsgBus:      msgBus,
		}
		return tbft.New(config)
	case consensuspb.ConsensusType_SOLO:
		return solo.New(chainID, id, signer, msgBus, chainConf)
	case consensuspb.ConsensusType_RAFT:
		config := raft.ConsensusRaftImplConfig{
			ChainID:        chainID,
			Singer:         signer,
			Ac:             ac,
			LedgerCache:    ledgerCache,
			BlockVerifier:  blockVerifier,
			BlockCommitter: blockCommitter,
			ChainConf:      chainConf,
			MsgBus:         msgBus,
		}
		return raft.New(config)
	default:
	}
	return nil, (fmt.Errorf("error consensusType: %s", consensusType))
}

// VerifyBlockSignatures verifies whether the signatures in block
// is qulified with the consensus algorithm. It should return nil
// error when verify successfully, and return corresponding error
// when failed.
func VerifyBlockSignatures(
	chainConf protocol.ChainConf,
	ac protocol.AccessControlProvider,
	store protocol.BlockchainStore,
	block *commonpb.Block,
) error {
	consensusType := chainConf.ChainConfig().Consensus.Type
	switch consensusType {
	case consensuspb.ConsensusType_TBFT:
		return tbft.VerifyBlockSignatures(chainConf, ac, block)
	case consensuspb.ConsensusType_RAFT:
		return raft.VerifyBlockSignatures(block)
	case consensuspb.ConsensusType_SOLO:
		fallthrough
	default:
	}
	return fmt.Errorf("error consensusType: %s", consensusType)
}