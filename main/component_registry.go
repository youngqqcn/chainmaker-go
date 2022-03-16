/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"chainmaker.org/chainmaker-go/module/consensus"
	"chainmaker.org/chainmaker-go/module/txpool"
	"chainmaker.org/chainmaker-go/module/vm"
	dpos "chainmaker.org/chainmaker/consensus-dpos/v2"
	maxbft "chainmaker.org/chainmaker/consensus-maxbft/v2"
	raft "chainmaker.org/chainmaker/consensus-raft/v2"
	solo "chainmaker.org/chainmaker/consensus-solo/v2"
	tbft "chainmaker.org/chainmaker/consensus-tbft/v2"
	utils "chainmaker.org/chainmaker/consensus-utils/v2"
	"chainmaker.org/chainmaker/localconf/v2"
	consensusPb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
	batch "chainmaker.org/chainmaker/txpool-batch/v2"
	single "chainmaker.org/chainmaker/txpool-single/v2"
	dockergo "chainmaker.org/chainmaker/vm-docker-go/v2"
	evm "chainmaker.org/chainmaker/vm-evm/v2"
	gasm "chainmaker.org/chainmaker/vm-gasm/v2"
	wasmer "chainmaker.org/chainmaker/vm-wasmer/v2"
	wxvm "chainmaker.org/chainmaker/vm-wxvm/v2"
)

func init() {
	// txPool
	txpool.RegisterTxPoolProvider(single.TxPoolType, single.NewTxPoolImpl)
	txpool.RegisterTxPoolProvider(batch.TxPoolType, batch.NewBatchTxPool)

	// vm
	vm.RegisterVmProvider(
		"GASM",
		func(chainId string, configs map[string]interface{}) (protocol.VmInstancesManager, error) {
			return &gasm.InstancesManager{}, nil
		})
	vm.RegisterVmProvider(
		"WASMER",
		func(chainId string, configs map[string]interface{}) (protocol.VmInstancesManager, error) {
			return wasmer.NewInstancesManager(chainId), nil
		})
	vm.RegisterVmProvider(
		"WXVM",
		func(chainId string, configs map[string]interface{}) (protocol.VmInstancesManager, error) {
			return &wxvm.InstancesManager{}, nil
		})
	vm.RegisterVmProvider(
		"EVM",
		func(chainId string, configs map[string]interface{}) (protocol.VmInstancesManager, error) {
			return &evm.InstancesManager{}, nil
		})

	vm.RegisterVmProvider(
		"DOCKERGO",
		func(chainId string, configs map[string]interface{}) (protocol.VmInstancesManager, error) {
			return dockergo.NewDockerManager(chainId, localconf.ChainMakerConfig.VMConfig), nil
		})

	// consensus
	consensus.RegisterConsensusProvider(
		consensusPb.ConsensusType_SOLO,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return solo.New(config)
		},
	)

	consensus.RegisterConsensusProvider(
		consensusPb.ConsensusType_DPOS,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			tbftEngine, err := tbft.New(config) // DPoS based in TBFT
			if err != nil {
				return nil, err
			}
			dposEngine := dpos.NewDPoSImpl(config, tbftEngine)
			return dposEngine, nil
		},
	)

	consensus.RegisterConsensusProvider(
		consensusPb.ConsensusType_RAFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return raft.New(config)
		},
	)

	consensus.RegisterConsensusProvider(
		consensusPb.ConsensusType_TBFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return tbft.New(config)
		},
	)

	consensus.RegisterConsensusProvider(
		consensusPb.ConsensusType_MAXBFT,
		func(config *utils.ConsensusImplConfig) (protocol.ConsensusEngine, error) {
			return maxbft.New(config)
		},
	)
}
