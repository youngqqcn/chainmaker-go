/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"fmt"
	"sync/atomic"

	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"

	"chainmaker.org/chainmaker/localconf/v2"
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
	case *ReceivedBlockInfos:
		pro.log.Debugf("receive [ReceivedBlockInfos] msg, start handle...", msg.Data)
		pro.handleReceivedBlockInfos(msg)
	case *ProcessBlockMsg:
		//pro.log.Debugf("receive [ProcessBlockMsg] msg, start handle...")
		return pro.handleProcessBlockMsg()
	case *DataDetection:
		pro.log.Debugf("receive [DataDetection] msg, start handle...")
		pro.handleDataDetection()
	}
	return nil, nil
}

func (pro *processor) handleReceivedBlockInfos(msg *ReceivedBlockInfos) {
	pro.log.Info("handleReceivedBlockInfos start")
	lastCommitBlockHeight := pro.lastCommitBlockHeight()
	switch ty := msg.Data.(type) {
	// block message implied in chainmaker version <= 2.1.x
	case *syncPb.SyncBlockBatch_BlockBatch:
		for _, blk := range ty.BlockBatch.Batches {
			if blk.Header.BlockHeight <= lastCommitBlockHeight {
				continue
			}
			if _, exist := pro.queue[blk.Header.BlockHeight]; !exist {
				pro.queue[blk.Header.BlockHeight] = blockWithPeerInfo{
					blk: blk, rwsets: nil, id: msg.from, withRWSets: msg.WithRwset,
				}
				pro.log.Debugf("received block [height: %d] from node [%s]",
					blk.Header.BlockHeight, msg.from)
				pro.log.Debugf("current length of processor queue is: [%d]", len(pro.queue))
			}
		}
	// block message implied in chainmaker version == 2.2.x
	case *syncPb.SyncBlockBatch_BlockinfoBatch:
		for _, blkinfo := range ty.BlockinfoBatch.Batch {
			if blkinfo.Block.Header.BlockHeight <= lastCommitBlockHeight {
				continue
			}
			if _, exist := pro.queue[blkinfo.Block.Header.BlockHeight]; !exist {
				pro.queue[blkinfo.Block.Header.BlockHeight] = blockWithPeerInfo{
					blk: blkinfo.Block, rwsets: blkinfo.RwsetList, id: msg.from, withRWSets: msg.WithRwset,
				}
				pro.log.Debugf("received block with rwsets [height: %d] from node [%s]",
					blkinfo.Block.Header.BlockHeight, msg.from)
				pro.log.Debugf("current length of processor queue is: [%d]", len(pro.queue))
			}
		}
	default:
		pro.log.Errorf("handleReceivedBlockInfos get unrecognized msg, type: %t", msg.Data)
	}
}

func (pro *processor) handleProcessBlockMsg() (queue.Item, error) {
	var (
		exist  bool
		info   blockWithPeerInfo
		status processedBlockStatus
	)
	pendingBlockHeight := pro.lastCommitBlockHeight() + 1
	if info, exist = pro.queue[pendingBlockHeight]; !exist {
		//pro.log.Debugf("block [%d] not find in queue.", pendingBlockHeight)
		return nil, nil
	}
	pro.log.Debugf("process block [height: %d] start, status [%d]", info.blk.Header.BlockHeight, status)
	if info.withRWSets && localconf.ChainMakerConfig.NodeConfig.FastSyncConfig.Enable {
		if status = pro.validateAndCommitBlockWithRwSets(info.blk, info.rwsets); status == ok || status == hasProcessed {
			pro.hasCommitBlock++
		}
	} else {
		if status = pro.validateAndCommitBlock(info.blk); status == ok || status == hasProcessed {
			pro.hasCommitBlock++
		}
	}
	delete(pro.queue, pendingBlockHeight)
	pro.log.Infof("process block [height: %d] success, status [%d]", info.blk.Header.BlockHeight, status)
	pro.log.DebugDynamic(pro.getServiceState)
	return &ProcessedBlockResp{
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
