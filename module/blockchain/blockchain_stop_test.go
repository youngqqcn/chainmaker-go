/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/
package blockchain

import "testing"

func TestStopOnRequirements(t *testing.T) {
	t.Log("TestStopOnRequirements")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}
		blockchain.StopOnRequirements()
	}
}

func TestStopNetService(t *testing.T) {
	t.Log("TestStopNetService")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}

		err = blockchain.stopNetService()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStopConsensus(t *testing.T) {
	t.Log("TestStopConsensus")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}
		err = blockchain.stopConsensus()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStopCoreEngine(t *testing.T) {
	t.Log("TestStopCoreEngine")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}
		err = blockchain.stopCoreEngine()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStopSyncService(t *testing.T) {
	t.Log("TestStopSyncService")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}
		err = blockchain.stopSyncService()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestStopTxPool(t *testing.T) {
	t.Log("TestStopTxPool")

	blockchainList := createBlockChain(t)

	for _, blockchain := range blockchainList {
		err := blockchain.Start()
		if err != nil {
			t.Log(err)
		}
		err = blockchain.stopTxPool()
		if err != nil {
			t.Log(err)
		}
	}
}
