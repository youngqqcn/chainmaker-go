/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/
package blockchain

import (
	"testing"
)

func TestInitSubscriber(t *testing.T) {
	t.Log("TestInitSubscriber")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initSubscriber()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitCache(t *testing.T) {
	t.Log("TestInitCache")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initCache()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitChainConf(t *testing.T) {
	t.Log("TestInitChainConf")

	blockchainList := createBlockChain(t)

	var err error
	for _, blockchain := range blockchainList {

		err = blockchain.initChainConf() // TODO 可以加上一些 conf 不为空的情况
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitSync(t *testing.T) {
	t.Log("TestInitSync")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initSync()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitNetService(t *testing.T) {
	t.Log("TestInitNetService")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initNetService()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitAC(t *testing.T) {
	t.Log("TestInitNetService")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initAC()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitTxPool(t *testing.T) {
	t.Log("TestInitTxPool")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initTxPool()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitVM(t *testing.T) {
	t.Log("TestInitVM")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {

		err := blockchain.initVM()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestGetChainNodesInfo(t *testing.T) {
	t.Log("TestGetChainNodesInfo")

	for i := 0; i < 6; i++ {
		chainNoInfo, err := (&soloChainNodesInfoProvider{}).GetChainNodesInfo()

		if err != nil {
			t.Log(err)
		}

		t.Log(chainNoInfo)
	}

}

func TestInitCore(t *testing.T) {
	t.Log("TestInitCore")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initCore()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitConsensus(t *testing.T) {
	t.Log("TestInitCore")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initConsensus()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitBaseModules(t *testing.T) {
	t.Log("TestInitBaseModules")

	blockchainList := createBlockChain(t)

	for k, blockchain := range blockchainList { // TODO 可以再加入一些条件

		if k < 1 {
			err := blockchain.initBaseModules([]map[string]func() error{
				// init Subscriber
				{moduleNameSubscriber: blockchain.initSubscriber},
				// init store module
				//{moduleNameStore: blockchain.initStore}, TODO
				// init ledger module
				{moduleNameLedger: blockchain.initCache},
				// init chain config , must latter than store module
				//{moduleNameChainConf: blockchain.initChainConf},
			})
			if err != nil {
				t.Log(err)
			}
		} else if k < 2 {
			err := blockchain.initBaseModules([]map[string]func() error{
				// init Subscriber
				//{moduleNameSubscriber: blockchain.initSubscriber},
				// init store module
				//{moduleNameStore: blockchain.initStore}, TODO
				// init ledger module
				{moduleNameLedger: blockchain.initCache},
				// init chain config , must latter than store module
				//{moduleNameChainConf: blockchain.initChainConf},
			})
			if err != nil {
				t.Log(err)
			}
		} else {
			err := blockchain.initBaseModules([]map[string]func() error{
				// init Subscriber
				{moduleNameSubscriber: blockchain.initSubscriber},
				// init store module
				//{moduleNameStore: blockchain.initStore}, TODO
				// init ledger module
				//{moduleNameLedger: blockchain.initCache},
				// init chain config , must latter than store module
				//{moduleNameChainConf: blockchain.initChainConf},
			})
			if err != nil {
				t.Log(err)
			}
		}

	}

}

// TODO
/*func TestInitStore(t *testing.T) {
	t.Log("TestInitStore")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initSubscriber()

		if err != nil {
			t.Log(err)
		}


		err = blockchain.initStore()

		if err != nil {
			t.Log(err)
		}
	}
}*/

/*func TestInit(t *testing.T) {
	t.Log("TestInit")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.Init()
		if err != nil {
			t.Log(err)
		}
	}
}*/
