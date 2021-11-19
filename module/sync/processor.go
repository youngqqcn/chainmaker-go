/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"chainmaker.org/chainmaker/localconf/v2"
	"fmt"
	"sync/atomic"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"

	"chainmaker.org/chainmaker/protocol/v2"

	"github.com/Workiva/go-datastructures/queue"
)

type verifyAndAddBlock interface {
	validateAndCommitBlock(block *commonPb.Block) processedBlockStatus
	validateAndCommitBlockWithRwSets(block *commonPb.Block, rwsets []*commonPb.TxRWSet) processedBlockStatus
}

type blockWithPeerInfo struct {
	id         string
	blk        *commonPb.Block
	withRWSets bool
	rwsets     []*commonPb.TxRWSet
}

type processor struct {
	queue          map[uint64]blockWithPeerInfo // Information about the blocks will be processed
	hasCommitBlock uint64                       // Number of blocks that have been commit

	log         protocol.Logger
	ledgerCache protocol.LedgerCache // Provides the latest chain state for the node
	verifyAndAddBlock
}

func newProcessor(verify verifyAndAddBlock, ledgerCache protocol.LedgerCache, log protocol.Logger) *processor {
	return &processor{
		ledgerCache:       ledgerCache,
		verifyAndAddBlock: verify,
		queue:             make(map[uint64]blockWithPeerInfo),
		log:               log,
	}
}

func (pro *processor) handler(event queue.Item) (queue.Item, error) {
	switch msg := event.(type) {
	case *ReceivedBlocks:
		pro.handleReceivedBlocks(msg)
	case *ReceivedBlocksWithRwSets:
		pro.handleReceivedBlocksWithRwSets(msg)
	case ProcessBlockMsg:
		return pro.handleProcessBlockMsg()
	//case ProcessBlockWithRwSetMsg:
	//	return pro.handleProcessBlockWithRwSetMsg()
	case DataDetection:
		pro.handleDataDetection()
	}
	return nil, nil
}

func (pro *processor) handleReceivedBlocks(msg *ReceivedBlocks) {
	pro.log.Info("handleReceivedBlocks start")
	lastCommitBlockHeight := pro.lastCommitBlockHeight()
	for _, blk := range msg.blks {
		if blk.Header.BlockHeight <= lastCommitBlockHeight {
			continue
		}
		if _, exist := pro.queue[blk.Header.BlockHeight]; !exist {
			pro.queue[blk.Header.BlockHeight] = blockWithPeerInfo{
				blk: blk, id: msg.from, withRWSets: false,
			}
			pro.log.Debugf("received block [height: %d] from node [%s]", blk.Header.BlockHeight, msg.from)
		}
	}
}

func (pro *processor) handleReceivedBlocksWithRwSets(msg *ReceivedBlocksWithRwSets) {
	pro.log.Info("handleReceivedBlocksWithRwSets start")
	lastCommitBlockHeight := pro.lastCommitBlockHeight()
	for _, blkinfo := range msg.blkinfos {
		if blkinfo.Block.Header.BlockHeight <= lastCommitBlockHeight {
			continue
		}
		if _, exist := pro.queue[blkinfo.Block.Header.BlockHeight]; !exist {
			pro.queue[blkinfo.Block.Header.BlockHeight] = blockWithPeerInfo{
				blk: blkinfo.Block, rwsets: blkinfo.RwsetList, id: msg.from, withRWSets: true,
			}
			pro.log.Debugf("received block [height: %d] from node [%s]", blkinfo.Block.Header.BlockHeight, msg.from)
		}
	}
}

func (pro *processor) handleProcessBlockMsg() (queue.Item, error) {
	var (
		exist  bool
		info   blockWithPeerInfo
		status processedBlockStatus
	)
	pendingBlockHeight := pro.lastCommitBlockHeight() + 1
	isFastSync := localconf.ChainMakerConfig.NodeConfig.FastSyncConfig.Enable
	if info, exist = pro.queue[pendingBlockHeight]; !exist {
		//pro.log.Debugf("block [%d] not find in queue.", pendingBlockHeight)
		return nil, nil
	}
	if info.withRWSets && isFastSync {
		if status = pro.validateAndCommitBlockWithRwSets(info.blk, info.rwsets); status == ok || status == hasProcessed {
			pro.hasCommitBlock++
		}
	} else {
		if status = pro.validateAndCommitBlock(info.blk); status == ok || status == hasProcessed {
			pro.hasCommitBlock++
		}
	}
	delete(pro.queue, pendingBlockHeight)
	pro.log.Infof("process block [height: %d], status [%d]", info.blk.Header.BlockHeight, status)
	return ProcessedBlockResp{
		status: status,
		height: info.blk.Header.BlockHeight,
		from:   info.id,
	}, nil
}

func (pro *processor) handleDataDetection() {
	pendingBlockHeight := pro.lastCommitBlockHeight() + 1
	for height := range pro.queue {
		if height < pendingBlockHeight {
			delete(pro.queue, height)
		}
	}
}

func (pro *processor) lastCommitBlockHeight() uint64 {
	return pro.ledgerCache.GetLastCommittedBlock().Header.BlockHeight
}

func (pro *processor) hasProcessedBlock() uint64 {
	return atomic.LoadUint64(&pro.hasCommitBlock)
}

func (pro *processor) getServiceState() string {
	return fmt.Sprintf("pendingBlockHeight: %d, queue num: %d", pro.lastCommitBlockHeight()+1, len(pro.queue))
}
