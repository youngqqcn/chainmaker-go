/*
 * Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package consensus

import (
	"fmt"

	dpos "chainmaker.org/chainmaker/consensus-dpos/v2"
	maxbft "chainmaker.org/chainmaker/consensus-maxbft/v2"
	raft "chainmaker.org/chainmaker/consensus-raft/v2"
	tbft "chainmaker.org/chainmaker/consensus-tbft/v2"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
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
	case consensuspb.ConsensusType_TBFT:
		// get validator list by module of tbft
		return tbft.VerifyBlockSignatures(chainConf, ac, block, store, tbft.GetValidatorList)
	case consensuspb.ConsensusType_DPOS:
		// get validator list by module of dpos
		return tbft.VerifyBlockSignatures(chainConf, ac, block, store, dpos.GetValidatorList)
	case consensuspb.ConsensusType_RAFT:
		return raft.VerifyBlockSignatures(block)
	case consensuspb.ConsensusType_MAXBFT:
		return maxbft.VerifyBlockSignatures(chainConf, ac, store, block, ledger)
	case consensuspb.ConsensusType_SOLO:
		return nil //for rebuild-dbs
	default:
	}
	return fmt.Errorf("error consensusType: %s", consensusType)
}
