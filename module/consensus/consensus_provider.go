package consensus

import (
	"chainmaker.org/chainmaker-go/consensus/implconfig"
	consensusPb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
)

type Provider func(config *implconfig.ConsensusImplConfig) (protocol.ConsensusEngine, error)

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
