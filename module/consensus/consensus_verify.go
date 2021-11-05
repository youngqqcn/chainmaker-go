/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consensus

import (
	"fmt"

	"chainmaker.org/chainmaker-go/consensus/chainedbft"

	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"

	"chainmaker.org/chainmaker-go/consensus/raft"
	"chainmaker.org/chainmaker-go/consensus/tbft"
	"chainmaker.org/chainmaker/protocol/v2"
)

// VerifyBlockSignatures verifies whether the signatures in block
// is qulified with the consensus algorithm. It should return nil
// error when verify successfully, and return corresponding error
// when failed.
func VerifyBlockSignatures(
	chainConf protocol.ChainConf,
	ac protocol.AccessControlProvider,
	store protocol.BlockchainStore,
	block *commonpb.Block,
	ledger protocol.LedgerCache,
) error {
	consensusType := chainConf.ChainConfig().Consensus.Type
	switch consensusType {
	case consensuspb.ConsensusType_TBFT, consensuspb.ConsensusType_DPOS:
		return tbft.VerifyBlockSignatures(chainConf, ac, block, store)
	case consensuspb.ConsensusType_RAFT:
		return raft.VerifyBlockSignatures(block)
	case consensuspb.ConsensusType_HOTSTUFF:
		return chainedbft.VerifyBlockSignatures(chainConf, ac, store, block, ledger)
	case consensuspb.ConsensusType_SOLO:
		fallthrough
	default:
	}
	return fmt.Errorf("error consensusType: %s", consensusType)
}
