/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"os"
	"time"

	commonErrors "chainmaker.org/chainmaker/common/v2/errors"
)

// Start all the modules.
func (bc *Blockchain) RebuildDbs() {
	bc.log.Infof("start RebuildDbs...")
	lastBlock, err := bc.oldStore.GetLastBlock()
	if err != nil {
		bc.log.Errorf("get lastblockerr(%s)", err.Error())
	} else {
		bc.log.Infof("lastBlock=%d", lastBlock.Header.BlockHeight)
	}
	var i uint64
	var preHash []byte
	for i = 1; i <= lastBlock.GetHeader().BlockHeight; i++ {
		block, err := bc.oldStore.GetBlock(uint64(i))
		if err != nil {
			bc.log.Errorf("get block %d err(%s)", i, err.Error())
		}
		bc.log.Infof("block %d hash is %x", i, block.GetHeader().BlockHash)
		bc.log.Infof("block %d prehash is %x", i, block.GetHeader().PreBlockHash)
		if preHash != nil && string(preHash) != string(block.GetHeader().PreBlockHash) {
			bc.log.Infof("\npreHash=%x\nprehash=%x", []byte(preHash), block.GetHeader().PreBlockHash)
			bc.log.Errorf("\nError!!!!\n")
		} else {
			bc.log.Infof("\npreHash=%x\nprehash=%x", []byte(preHash), block.GetHeader().PreBlockHash)
		}
		preHash = block.GetHeader().BlockHash

		////bc.msgBus
		//bc.msgBus.Publish(msgbus.RebuildVerifyBlock, block)
		//time.Sleep(500 * time.Millisecond)
		//bc.msgBus.Publish(msgbus.RebuildCommitBlock, block)
		////bc.msgBus.Publish(msgbus.BlockInfo, &common.BlockInfo{Block: block})
		//time.Sleep(500 * time.Millisecond)

		if err := bc.coreEngine.GetBlockVerifier().VerifyBlock(block, -1); err != nil {
			if err == commonErrors.ErrBlockHadBeenCommited {
				bc.log.Errorf("the block: %d has been committed in the blockChainStore ", block.Header.BlockHeight)
			} else {
				bc.log.Infof("block[%d] verify success.", block.Header.BlockHeight)
			}
		}

		//time.Sleep(500*time.Millisecond)
		if err := bc.coreEngine.GetBlockCommitter().AddBlock(block); err != nil {
			if err == commonErrors.ErrBlockHadBeenCommited {
				bc.log.Errorf("the block: %d has been committed in the blockChainStore ", block.Header.BlockHeight)
			} else {
				bc.log.Infof("block[%d] commit success.", block.Header.BlockHeight)
			}

		}
		time.Sleep(500 * time.Millisecond)

	}
	bc.log.Infof("###########################")
	bc.log.Infof("###rebuild-dbs finished!###")
	bc.log.Infof("###########################")
	bc.Stop()
	os.Exit(0)
}
