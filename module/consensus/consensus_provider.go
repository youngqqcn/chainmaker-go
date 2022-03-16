/*

Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package consensus

import (
	utils "chainmaker.org/chainmaker/consensus-utils/v2"
	consensusPb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
)

type Provider func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error)

var consensusProviders = make(map[consensusPb.ConsensusType]Provider)

func RegisterConsensusProvider(t consensusPb.ConsensusType, f Provider) {
	consensusProviders[t] = f
}

func GetConsensusProvider(t consensusPb.ConsensusType) Provider {
	provider, ok := consensusProviders[t]
	if !ok {
		return nil
	}
	return provider
}
