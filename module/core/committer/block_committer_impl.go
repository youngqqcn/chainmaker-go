/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package committer

import (
	"bytes"
	"chainmaker.org/chainmaker-go/chainconf"
	commonErrors "chainmaker.org/chainmaker-go/common/errors"
	"chainmaker.org/chainmaker-go/common/msgbus"
	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker-go/monitor"
	commonpb "chainmaker.org/chainmaker-go/pb/protogo/common"
	"chainmaker.org/chainmaker-go/protocol"
	"chainmaker.org/chainmaker-go/subscriber"
	"chainmaker.org/chainmaker-go/utils"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// BlockCommitterImpl implements BlockCommitter interface.
// To commit a block after it is confirmed by consensus module.
type BlockCommitterImpl struct {
	chainId string // chain id, to identity this chain
	// Store is a block store that will only fetch data locally
	blockchainStore protocol.BlockchainStore // blockchain store
	snapshotManager protocol.SnapshotManager // snapshot manager
	txPool          protocol.TxPool          // transaction pool
	chainConf       protocol.ChainConf       // chain config

	ledgerCache           protocol.LedgerCache        // ledger cache
	proposalCache         protocol.ProposalCache      // proposal cache
	log                   *logger.CMLogger            // logger
	msgBus                msgbus.MessageBus           // message bus
	mu                    sync.Mutex                  // lock, to avoid concurrent block commit
	subscriber            *subscriber.EventSubscriber // subscriber
	verifier              protocol.BlockVerifier      // block verifier
	metricBlockSize       *prometheus.HistogramVec    // metric block size
	metricBlockCounter    *prometheus.CounterVec      // metric block counter
	metricTxCounter       *prometheus.CounterVec      // metric transaction counter
	metricBlockCommitTime *prometheus.HistogramVec    // metric block commit time
}

type BlockCommitterConfig struct {
	ChainId         string
	BlockchainStore protocol.BlockchainStore
	SnapshotManager protocol.SnapshotManager
	TxPool          protocol.TxPool
	LedgerCache     protocol.LedgerCache
	ProposedCache   protocol.ProposalCache
	ChainConf       protocol.ChainConf
	MsgBus          msgbus.MessageBus
	Subscriber      *subscriber.EventSubscriber
	Verifier        protocol.BlockVerifier
}

func NewBlockCommitter(config BlockCommitterConfig) (protocol.BlockCommitter, error) {
	blockchain := &BlockCommitterImpl{
		chainId:         config.ChainId,
		blockchainStore: config.BlockchainStore,
		snapshotManager: config.SnapshotManager,
		txPool:          config.TxPool,
		ledgerCache:     config.LedgerCache,
		proposalCache:   config.ProposedCache,
		log:             logger.GetLoggerByChain(logger.MODULE_CORE, config.ChainId),
		chainConf:       config.ChainConf,
		msgBus:          config.MsgBus,
		subscriber:      config.Subscriber,
		verifier:        config.Verifier,
	}

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		blockchain.metricBlockSize = monitor.NewHistogramVec(monitor.SUBSYSTEM_CORE_COMMITTER, monitor.MetricBlockSize,
			monitor.HelpCurrentBlockSizeMetric, prometheus.ExponentialBuckets(1024, 2, 12), monitor.ChainId)

		blockchain.metricBlockCounter = monitor.NewCounterVec(monitor.SUBSYSTEM_CORE_COMMITTER, monitor.MetricBlockCounter,
			monitor.HelpBlockCountsMetric, monitor.ChainId)

		blockchain.metricTxCounter = monitor.NewCounterVec(monitor.SUBSYSTEM_CORE_COMMITTER, monitor.MetricTxCounter,
			monitor.HelpTxCountsMetric, monitor.ChainId)

		blockchain.metricBlockCommitTime = monitor.NewHistogramVec(monitor.SUBSYSTEM_CORE_COMMITTER, monitor.MetricBlockCommitTime,
			monitor.HelpBlockCommitTimeMetric, []float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 10}, monitor.ChainId)
	}

	return blockchain, nil
}

func (chain *BlockCommitterImpl) isBlockLegal(blk *commonpb.Block) error {
	lastBlock := chain.ledgerCache.GetLastCommittedBlock()
	if lastBlock == nil {
		// 获取上一区块
		// 首次进入，从DB获取最新区块
		return fmt.Errorf("get last block == nil ")
	}

	if lastBlock.Header.BlockHeight >= blk.Header.BlockHeight {
		return commonErrors.ErrBlockHadBeenCommited
	}
	// block height verify
	if blk.Header.BlockHeight != lastBlock.Header.BlockHeight+1 {
		return fmt.Errorf("isBlockLegal() failed: Height is less than chaintip")
	}
	// block pre hash verify
	if !bytes.Equal(blk.Header.PreBlockHash, lastBlock.Header.BlockHash) {
		return fmt.Errorf("isBlockLegal() failed: PrevHash invalid (%x != %x)",
			blk.Header.PreBlockHash, lastBlock.Header.BlockHash)
	}

	blkHash, err := utils.CalcBlockHash(chain.chainConf.ChainConfig().Crypto.Hash, blk)
	if err != nil || !bytes.Equal(blk.Header.BlockHash, blkHash) {
		return fmt.Errorf("isBlockLegal() failed: BlockHash invalid (%x != %x)",
			blkHash, blk.Header.BlockHash)
	}

	return nil
}

func (chain *BlockCommitterImpl) AddBlock(block *commonpb.Block) error {
	startTick := utils.CurrentTimeMillisSeconds()
	chain.log.Debugf("add block(%d,%x)=(%x,%d,%d)",
		block.Header.BlockHeight, block.Header.BlockHash, block.Header.PreBlockHash, block.Header.TxCount, len(block.Txs))
	chain.mu.Lock()
	defer chain.mu.Unlock()
	var err error

	height := block.Header.BlockHeight
	if err = chain.isBlockLegal(block); err != nil {
		chain.log.Errorf("block illegal [%d](hash:%x), %s", height, block.Header.BlockHash, err)
		return err
	}
	lastProposed, rwSetMap := chain.proposalCache.GetProposedBlock(block)
	if err = chain.checkLastProposedBlock(block, lastProposed, err, height, rwSetMap); err != nil {
		return err
	}

	// record block
	rwSet := chain.rearrangeRWSet(block, rwSetMap)

	checkLasts := utils.CurrentTimeMillisSeconds() - startTick
	startDBTick := utils.CurrentTimeMillisSeconds()
	if err = chain.blockchainStore.PutBlock(block, rwSet); err != nil {
		// if put db error, then panic
		chain.log.Error(err)
		panic(err)
	}
	dbLasts := utils.CurrentTimeMillisSeconds() - startDBTick

	// clear snapshot
	startSnapshotTick := utils.CurrentTimeMillisSeconds()
	if err = chain.snapshotManager.NotifyBlockCommitted(block); err != nil {
		err = fmt.Errorf("notify snapshot error [%d](hash:%x)",
			lastProposed.Header.BlockHeight, lastProposed.Header.BlockHash)
		chain.log.Error(err)
		return err
	}
	snapshotLasts := utils.CurrentTimeMillisSeconds() - startSnapshotTick

	// notify chainConf to update config when config block committed
	startConfTick := utils.CurrentTimeMillisSeconds()
	if err = chain.notifyChainConf(block, err); err != nil {
		return err
	}
	confLasts := utils.CurrentTimeMillisSeconds() - startConfTick

	// Remove txs from txpool. Remove will invoke proposeSignal from txpool if pool size > txcount
	startPoolTick := utils.CurrentTimeMillisSeconds()
	txRetry := chain.syncWithTxPool(block, height)
	chain.log.Infof("remove txs[%d] and retry txs[%d] in add block", len(block.Txs), len(txRetry))
	chain.txPool.RetryAndRemoveTxs(txRetry, block.Txs)
	poolLasts := utils.CurrentTimeMillisSeconds() - startPoolTick

	startOtherTick := utils.CurrentTimeMillisSeconds()
	chain.ledgerCache.SetLastCommittedBlock(block)
	chain.proposalCache.ClearProposedBlockAt(height)
	bi := &commonpb.BlockInfo{
		Block:     block,
		RwsetList: rwSet,
	}
	// synchronize new block height to consensus and sync module
	chain.msgBus.Publish(msgbus.BlockInfo, bi)

	if err = chain.monitorCommit(bi); err != nil {
		return err
	}

	otherLasts := utils.CurrentTimeMillisSeconds() - startOtherTick
	elapsed := utils.CurrentTimeMillisSeconds() - startTick
	chain.log.Infof("commit block [%d](count:%d,hash:%x), time used(check:%d,db:%d,ss:%d,conf:%d,pool:%d,other:%d,total:%d)",
		height, block.Header.TxCount, block.Header.BlockHash, checkLasts, dbLasts, snapshotLasts, confLasts, poolLasts, otherLasts, elapsed)
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		chain.metricBlockCommitTime.WithLabelValues(chain.chainId).Observe(float64(elapsed) / 1000)
	}
	return nil
}

func (chain *BlockCommitterImpl) monitorCommit(bi *commonpb.BlockInfo) error {
	if !localconf.ChainMakerConfig.MonitorConfig.Enabled {
		return nil
	}
	raw, err := proto.Marshal(bi)
	if err != nil {
		chain.log.Errorw("marshal BlockInfo failed", "err", err)
		return err
	}
	(*chain.metricBlockSize).WithLabelValues(bi.Block.Header.ChainId).Observe(float64(len(raw)))
	(*chain.metricBlockCounter).WithLabelValues(bi.Block.Header.ChainId).Inc()
	(*chain.metricTxCounter).WithLabelValues(bi.Block.Header.ChainId).Add(float64(bi.Block.Header.TxCount))
	return nil
}

func (chain *BlockCommitterImpl) syncWithTxPool(block *commonpb.Block, height int64) []*commonpb.Transaction {
	proposedBlocks := chain.proposalCache.GetProposedBlocksAt(height)
	txRetry := make([]*commonpb.Transaction, 0, localconf.ChainMakerConfig.TxPoolConfig.BatchMaxSize)
	batchMap := make(map[string]interface{})
	for _, b := range proposedBlocks {
		if bytes.Equal(b.Header.BlockHash, block.Header.BlockHash) {
			continue
		}
		if len(b.Txs) == 0 {
			continue
		}
		if _, ok := batchMap[b.Txs[0].Header.TxId]; ok {
			// make sure no redundant batch in txRetry
			continue
		}
		if len(block.Txs) > 0 && b.Txs[0].Header.TxId == block.Txs[0].Header.TxId {
			continue
		}
		batchMap[b.Txs[0].Header.TxId] = "exist"
		txRetry = append(txRetry, b.Txs...)
	}
	return txRetry
}

func (chain *BlockCommitterImpl) notifyChainConf(block *commonpb.Block, err error) error {
	if block != nil && block.GetTxs() != nil && len(block.GetTxs()) > 0 {
		tx := block.GetTxs()[0]
		if _, ok := chainconf.IsNativeTx(tx); ok {
			if err = chain.chainConf.CompleteBlock(block); err != nil {
				chain.log.Errorf("chainconf block complete, %s", err)
				return err
			}
		}
	}
	return nil
}

func (chain *BlockCommitterImpl) rearrangeRWSet(block *commonpb.Block, rwSetMap map[string]*commonpb.TxRWSet) []*commonpb.TxRWSet {
	rwSet := make([]*commonpb.TxRWSet, 0)
	if rwSetMap == nil {
		return rwSet
	}
	for _, tx := range block.Txs {
		if set, ok := rwSetMap[tx.Header.TxId]; ok {
			rwSet = append(rwSet, set)
		}
	}
	return rwSet
}

func (chain *BlockCommitterImpl) checkLastProposedBlock(block *commonpb.Block, lastProposed *commonpb.Block,
	err error, height int64, rwSetMap map[string]*commonpb.TxRWSet) error {
	if lastProposed != nil {
		return nil
	}
	err = chain.verifier.VerifyBlock(block, protocol.SYNC_VERIFY)
	if err != nil {
		chain.log.Error("block verify failed [%d](hash:%x), %s", height, block.Header.BlockHash, err)
		return err
	}
	lastProposed, rwSetMap = chain.proposalCache.GetProposedBlock(block)
	if lastProposed == nil {
		chain.log.Error("block not verified [%d](hash:%x)", height, block.Header.BlockHash)
		return fmt.Errorf("block not verified [%d](hash:%x)", height, block.Header.BlockHash)
	}
	return nil
}