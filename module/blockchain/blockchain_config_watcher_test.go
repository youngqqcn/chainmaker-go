/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/
package blockchain

import (
	"testing"
)

func TestModule(t *testing.T) {
	t.Log("TestModule")
	blockChainList := createBlockChain(t)

	for _, blockChain := range blockChainList {
		t.Log(blockChain.Module())
	}
}

// TODO
/*func TestWatch(t *testing.T) {
	t.Log("TestWatch")

	blockChainList := createBlockChain(t)
	for _, blockChain := range blockChainList {
		err := blockChain.Watch(blockChain.chainConf.ChainConfig())
		if err != nil {
			t.Log(err)
		}
	}
}*/
