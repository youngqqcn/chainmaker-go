package dpos

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
)

type NilDPoSImpl struct{}

func NewNilDPoSImpl() *NilDPoSImpl {
	return &NilDPoSImpl{}
}

func (n NilDPoSImpl) CreateDPoSRWSet(preBlkHash []byte, proposedBlock *consensus.ProposalBlock) error {
	return nil
}

func (n NilDPoSImpl) VerifyConsensusArgs(block *common.Block, blockTxRwSet map[string]*common.TxRWSet) error {
	return nil
}

func (n NilDPoSImpl) GetValidators() ([]string, error) {
	return nil, nil
}
