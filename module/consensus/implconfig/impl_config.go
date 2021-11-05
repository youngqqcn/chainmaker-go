package implconfig

import (
	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/protocol/v2"
)

type ConsensusImplConfig struct {
	ChainId       string
	NodeId        string
	Ac            protocol.AccessControlProvider
	Core          protocol.CoreEngine
	ChainConf     protocol.ChainConf
	Signer        protocol.SigningMember
	Store         protocol.BlockchainStore
	LedgerCache   protocol.LedgerCache
	ProposalCache protocol.ProposalCache
	MsgBus        msgbus.MessageBus
}

