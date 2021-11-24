/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/
package blockchain

import "testing"

func TestStartNetService(t *testing.T) {
	t.Log("TestStartNetService")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startNetService()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStartConsensus(t *testing.T) {
	t.Log("TestStartConsensus")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startConsensus()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStartCoreEngine(t *testing.T) {
	t.Log("TestStartCoreEngine")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startCoreEngine()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStartSyncService(t *testing.T) {
	t.Log("TestStartSyncService")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startSyncService()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStartTxPool(t *testing.T) {
	t.Log("TestStartTxPool")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startTxPool()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestIsModuleStartUp(t *testing.T) {
	t.Log("TestIsModuleStartUp")

	blockchainList := createBlockChain(t)

	moduleNameSlice := []string{
		moduleNameSubscriber,
		moduleNameStore,
		moduleNameLedger,
		moduleNameChainConf,
		moduleNameAccessControl,
		moduleNameNetService,
		moduleNameVM,
		moduleNameTxPool,
		moduleNameCore,
		moduleNameConsensus,
		moduleNameSync,
	}
	for _, moduleName := range moduleNameSlice {
		res := blockchainList[0].isModuleStartUp(moduleName)
		t.Log(res)
	}
}
