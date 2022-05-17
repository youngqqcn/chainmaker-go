/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package proposer

import (
	"bytes"
	"sync"
	"time"

	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"

	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/common/v2/monitor"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	pbac "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/pb-go/v2/consensus/maxbft"
	txpoolpb "chainmaker.org/chainmaker/pb-go/v2/txpool"
	"github.com/prometheus/client_golang/prometheus"
)

// BlockProposerImpl implements BlockProposer interface.
// In charge of propose a new block.
type BlockProposerImpl struct {
	chainId string // chain id, to identity this chain

	txPool          protocol.TxPool          // tx pool provides tx batch
	txScheduler     protocol.TxScheduler     // scheduler orders tx batch into DAG form and returns a block
	snapshotManager protocol.SnapshotManager // snapshot manager
	identity        protocol.SigningMember   // identity manager
	ledgerCache     protocol.LedgerCache     // ledger cache
	msgBus          msgbus.MessageBus        // channel to give out proposed block
	ac              protocol.AccessControlProvider
	blockchainStore protocol.BlockchainStore
	txFilter        protocol.TxFilter // Verify the transaction rules with TxFilter

	isProposer   bool        // whether current node can propose block now
	idle         bool        // whether current node is proposing or not
	proposeTimer *time.Timer // timer controls the proposing periods

	canProposeC   chan bool                   // channel to handle propose status change from consensus module
	txPoolSignalC chan *txpoolpb.TxPoolSignal // channel to handle propose signal from tx pool
	exitC         chan bool                   // channel to stop proposing loop
	proposalCache protocol.ProposalCache

	chainConf protocol.ChainConf // chain config

	idleMu         sync.Mutex   // for proposeBlock reentrant lock
	statusMu       sync.Mutex   // for propose status change lock
	proposerMu     sync.RWMutex // for isProposer lock, avoid race
	log            protocol.Logger
	finishProposeC chan bool // channel to receive signal to yield propose block

	metricBlockPackageTime *prometheus.HistogramVec
	proposer               *pbac.Member

	blockBuilder *common.BlockBuilder
	storeHelper  conf.StoreHelper
}

type BlockProposerConfig struct {
	ChainId         string
	TxPool          protocol.TxPool
	SnapshotManager protocol.SnapshotManager
	MsgBus          msgbus.MessageBus
	Identity        protocol.SigningMember
	LedgerCache     protocol.LedgerCache
	TxScheduler     protocol.TxScheduler
	ProposalCache   protocol.ProposalCache
	ChainConf       protocol.ChainConf
	AC              protocol.AccessControlProvider
	BlockchainStore protocol.BlockchainStore
	StoreHelper     conf.StoreHelper
	TxFilter        protocol.TxFilter
}

const (
	DEFAULTDURATION = 1000 // default proposal duration, millis seconds
)

func NewBlockProposer(config BlockProposerConfig, log protocol.Logger) (protocol.BlockProposer, error) {
	blockProposerImpl := &BlockProposerImpl{
		chainId:         config.ChainId,
		isProposer:      false, // not proposer when initialized
		idle:            true,
		msgBus:          config.MsgBus,
		blockchainStore: config.BlockchainStore,
		canProposeC:     make(chan bool),
		txPoolSignalC:   make(chan *txpoolpb.TxPoolSignal),
		exitC:           make(chan bool),
		txPool:          config.TxPool,
		snapshotManager: config.SnapshotManager,
		txScheduler:     config.TxScheduler,
		identity:        config.Identity,
		ledgerCache:     config.LedgerCache,
		proposalCache:   config.ProposalCache,
		chainConf:       config.ChainConf,
		ac:              config.AC,
		log:             log,
		finishProposeC:  make(chan bool),
		storeHelper:     config.StoreHelper,
		txFilter:        config.TxFilter,
	}

	var err error
	blockProposerImpl.proposer, err = blockProposerImpl.identity.GetMember()
	if err != nil {
		blockProposerImpl.log.Warnf("identity serialize failed, %s", err)
		return nil, err
	}

	// start propose timer
	blockProposerImpl.proposeTimer = time.NewTimer(blockProposerImpl.getDuration())
	if !blockProposerImpl.isSelfProposer() {
		blockProposerImpl.proposeTimer.Stop()
	}

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		blockProposerImpl.metricBlockPackageTime = monitor.NewHistogramVec(
			monitor.SUBSYSTEM_CORE_PROPOSER,
			"metric_block_package_time",
			"block package time metric",
			[]float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 10},
			"chainId",
		)
	}

	bbConf := &common.BlockBuilderConf{
		ChainId:         blockProposerImpl.chainId,
		TxPool:          blockProposerImpl.txPool,
		TxScheduler:     blockProposerImpl.txScheduler,
		SnapshotManager: blockProposerImpl.snapshotManager,
		Identity:        blockProposerImpl.identity,
		LedgerCache:     blockProposerImpl.ledgerCache,
		ProposalCache:   blockProposerImpl.proposalCache,
		ChainConf:       blockProposerImpl.chainConf,
		Log:             blockProposerImpl.log,
		StoreHelper:     blockProposerImpl.storeHelper,
	}

	blockProposerImpl.blockBuilder = common.NewBlockBuilder(bbConf)

	return blockProposerImpl, nil
}

// Start, start proposer
func (bp *BlockProposerImpl) Start() error {
	defer bp.log.Info("block proposer starts")

	go bp.startProposingLoop()

	return nil
}

// Stop, stop proposing loop
func (bp *BlockProposerImpl) Stop() error {
	defer bp.log.Infof("block proposer stoped")
	bp.exitC <- true
	return nil
}

// Start, start proposing loop
func (bp *BlockProposerImpl) startProposingLoop() {
	for {
		select {
		case <-bp.proposeTimer.C:
			if !bp.isSelfProposer() {
				break
			}
			go bp.proposeBlock()

		case signal := <-bp.txPoolSignalC:
			if !bp.isSelfProposer() {
				break
			}
			if signal.SignalType != txpoolpb.SignalType_BLOCK_PROPOSE {
				break
			}
			go bp.proposeBlock()

		case <-bp.exitC:
			bp.proposeTimer.Stop()
			bp.log.Info("block proposer loop stoped")
			return
		}
	}
}

/*
 * shouldProposeByBFT, check if node should propose new block
 * Only for *BFT consensus
 * if node is proposer, and node is not propose right now, and last proposed block is committed, then return true
 */
func (bp *BlockProposerImpl) shouldProposeByBFT(height uint64) bool {
	if !bp.isIdle() {
		// concurrent control, proposer is proposing now
		bp.log.Debugf("proposer is busy, not propose [%d] ", height)
		return false
	}
	committedBlock := bp.ledgerCache.GetLastCommittedBlock()
	if committedBlock == nil {
		bp.log.Errorf("no committed block found")
		return false
	}
	currentHeight := committedBlock.Header.BlockHeight
	// proposing height must higher than current height
	return currentHeight+1 == height
}

// proposeBlock, to check if proposer can propose block right now
// if so, start proposing
func (bp *BlockProposerImpl) proposeBlock() {
	defer func() {
		if bp.isSelfProposer() {
			bp.proposeTimer.Reset(bp.getDuration())
		}
	}()
	lastBlock := bp.ledgerCache.GetLastCommittedBlock()
	proposingHeight := lastBlock.Header.BlockHeight + 1
	if !bp.shouldProposeByBFT(proposingHeight) {
		return
	}
	if !bp.setNotIdle() {
		bp.log.Infof("concurrent propose block [%d], yield!", proposingHeight)
		return
	}
	defer bp.setIdle()

	go bp.proposing(proposingHeight, lastBlock.Header.BlockHash)
	// #DEBUG MODE#
	if localconf.ChainMakerConfig.DebugConfig.IsHaltPropose {
		go func() {
			bp.OnReceiveYieldProposeSignal(true)
		}()
	}

	<-bp.finishProposeC
}

// proposing, propose a block in new height
func (bp *BlockProposerImpl) proposing(height uint64, preHash []byte) *commonpb.Block {
	startTick := utils.CurrentTimeMillisSeconds()
	defer bp.yieldProposing()

	selfProposedBlock := bp.proposalCache.GetSelfProposedBlockAt(height)
	if selfProposedBlock != nil {
		if bytes.Equal(selfProposedBlock.Header.PreBlockHash, preHash) {
			// Repeat propose block if node has proposed before at the same height
			bp.proposalCache.SetProposedAt(height)
			_, txsRwSet, _ := bp.proposalCache.GetProposedBlock(selfProposedBlock)
			bp.msgBus.Publish(msgbus.ProposedBlock, &consensuspb.ProposalBlock{Block: selfProposedBlock, TxsRwSet: txsRwSet})
			bp.log.Infof("proposer success repeat [%d](txs:%d,hash:%x)",
				selfProposedBlock.Header.BlockHeight, selfProposedBlock.Header.TxCount, selfProposedBlock.Header.BlockHash)
			return nil
		}
		bp.proposalCache.ClearTheBlock(selfProposedBlock)
		// Note: It is not possible to re-add the transactions in the deleted block to txpool; because some transactions may
		// be included in other blocks to be confirmed, and it is impossible to quickly exclude these pending transactions
		// that have been entered into the block. Comprehensive considerations, directly discard this block is the optimal
		// choice. This processing method may only cause partial transaction loss at the current node, but it can be solved
		// by rebroadcasting on the client side.
		bp.txPool.RetryAndRemoveTxs(nil, selfProposedBlock.Txs)

	}
	var (
		fetchLasts          int64
		filterValidateLasts int64
		fetchTotalLasts     int64 // The total time consuming
		totalTimes          int   // loop count
		fetchBatch          []*commonpb.Transaction
	)
	// 根据TxFilter时间规则过滤交易，如果剩余的交易为0，则再次从交易池拉取交易，重复执行
	// The transaction is filtered according to txFilter time rule. If the remaining transaction is 0, the transaction
	// is pulled from the trading pool again and executed repeatedly
	fetchTotalFirst := utils.CurrentTimeMillisSeconds()
	for {
		totalTimes++
		// retrieve tx batch from tx pool
		fetchFirst := utils.CurrentTimeMillisSeconds()
		fetchBatch = bp.txPool.FetchTxBatch(height)
		fetchLasts += utils.CurrentTimeMillisSeconds() - fetchFirst
		bp.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("begin proposing block[%d], fetch tx num[%d]",
			height, len(fetchBatch)))
		if len(fetchBatch) == 0 {
			bp.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("no txs in tx pool, proposing block stoped"))
			return nil
		}
		// validate txFilter rules
		filterValidateFirst := utils.CurrentTimeMillisSeconds()
		removeTxs, remainTxs := common.ValidateTxRules(bp.txFilter, fetchBatch)
		filterValidateLasts += utils.CurrentTimeMillisSeconds() - filterValidateFirst
		if len(removeTxs) > 0 {
			// remove
			bp.txPool.RetryAndRemoveTxs(nil, removeTxs)
			bp.log.Warnf("remove the overtime transactions, total:%d, remain:%d, remove:%d",
				len(fetchBatch), len(remainTxs), len(removeTxs))
		}
		if len(remainTxs) > 0 {
			// 剩余交易大于0则跳出循环
			fetchBatch = remainTxs
			break
		}
	}
	fetchTotalLasts = utils.CurrentTimeMillisSeconds() - fetchTotalFirst
	if !utils.CanProposeEmptyBlock(bp.chainConf.ChainConfig().Consensus.Type) && len(fetchBatch) == 0 {
		// can not propose empty block and tx batch is empty, then yield proposing.
		bp.log.Debugf("no txs in tx pool, proposing block stoped")
		bp.txPool.RetryAndRemoveTxs(nil, fetchBatch)
		return nil
	}

	txCapacity := int(bp.chainConf.ChainConfig().Block.BlockTxCapacity)
	if len(fetchBatch) > txCapacity {
		// check if checkedBatch > txCapacity, if so, strict block tx count according to  config,
		// and put other txs back to txpool.
		txRetry := fetchBatch[txCapacity:]
		fetchBatch = fetchBatch[:txCapacity]
		bp.txPool.RetryAndRemoveTxs(txRetry, nil)
		bp.log.Warnf("txbatch oversize expect <= %d, got %d", txCapacity, len(fetchBatch))
	}

	block, timeLasts, err := bp.generateNewBlock(height, preHash, fetchBatch)
	if err != nil {
		bp.log.Warnf("generate new block failed, %s", err.Error())
		// rollback sql
		if sqlErr := bp.storeHelper.RollBack(block, bp.blockchainStore); sqlErr != nil {
			bp.log.Errorf("block [%d] rollback sql failed: %s", block.Header.BlockHeight, sqlErr)
		}
		bp.txPool.RetryAndRemoveTxs(fetchBatch, nil) // put txs back to txpool
		return nil
	}
	_, txsRwSet, _ := bp.proposalCache.GetProposedBlock(block)

	newBlock := new(commonpb.Block)
	if common.IfOpenConsensusMessageTurbo(bp.chainConf) {
		newBlock.Header = block.Header
		newBlock.Dag = block.Dag
		newTxs := make([]*commonpb.Transaction, len(block.Txs))
		for i := range block.Txs {
			newPayload := &commonpb.Payload{
				TxId: block.Txs[i].Payload.TxId,
			}

			newTxs[i] = &commonpb.Transaction{
				Payload:   newPayload,
				Sender:    block.Txs[i].Sender,
				Endorsers: block.Txs[i].Endorsers,
				Result:    block.Txs[i].Result,
			}
		}
		newBlock.Txs = newTxs
		bp.log.Debugf("turn on consensus message turbo, block[%d]", newBlock.Header.BlockHeight)
	} else {
		newBlock = block
	}

	bp.msgBus.Publish(msgbus.ProposedBlock, &consensuspb.ProposalBlock{Block: newBlock, TxsRwSet: txsRwSet})
	elapsed := utils.CurrentTimeMillisSeconds() - startTick
	bp.log.Infof("proposer success [%d](txs:%d), fetch(times:%v,fetch:%v,filter:%v,total:%d), vm:%v, total:%d",
		block.Header.BlockHeight, block.Header.TxCount, totalTimes, fetchLasts, filterValidateLasts, fetchTotalLasts,
		timeLasts, elapsed)
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		bp.metricBlockPackageTime.WithLabelValues(bp.chainId).Observe(float64(elapsed) / 1000)
	}
	return block
}

// OnReceiveTxPoolSignal, receive txpool signal and deliver to chan txpool signal
func (bp *BlockProposerImpl) OnReceiveTxPoolSignal(txPoolSignal *txpoolpb.TxPoolSignal) {
	bp.txPoolSignalC <- txPoolSignal
}

/*
 * OnReceiveProposeStatusChange, to update isProposer status when received proposeStatus from consensus
 * if node is proposer, then reset the timer, otherwise stop the timer
 */
func (bp *BlockProposerImpl) OnReceiveProposeStatusChange(proposeStatus bool) {
	bp.log.Debugf("OnReceiveProposeStatusChange(%t)", proposeStatus)
	bp.statusMu.Lock()
	defer bp.statusMu.Unlock()
	if proposeStatus == bp.isSelfProposer() {
		// 状态一致，忽略
		return
	}
	height, _ := bp.ledgerCache.CurrentHeight()
	bp.proposalCache.ResetProposedAt(height + 1) // proposer status changed, reset this round proposed status
	bp.setIsSelfProposer(proposeStatus)
	if !bp.isSelfProposer() {
		bp.yieldProposing() // try to yield if proposer self is proposing right now.
		bp.log.Debug("current node is not proposer ")
		return
	}
	bp.proposeTimer.Reset(bp.getDuration())
	bp.log.Debugf("current node is proposer, timeout period is %v", bp.getDuration())

}

// OnReceiveMaxBFTProposal, to check if this proposer should propose a new block
// Only for maxbft consensus
func (bp *BlockProposerImpl) OnReceiveMaxBFTProposal(proposal *maxbft.BuildProposal) {
	proposingHeight := proposal.Height
	preHash := proposal.PreHash
	if !bp.shouldProposeByMaxBFT(proposingHeight, preHash) {
		bp.log.Infof("not a legal proposal request [%d](%x)", proposingHeight, preHash)
		return
	}

	if !bp.setNotIdle() {
		bp.log.Warnf("concurrent propose block [%d](%x), yield!", proposingHeight, preHash)
		return
	}
	defer bp.setIdle()

	bp.log.Infof("trigger proposal from maxBFT, height[%d]", proposal.Height)
	go bp.proposing(proposingHeight, preHash)
	<-bp.finishProposeC
}

// OnReceiveYieldProposeSignal, receive yield propose signal
func (bp *BlockProposerImpl) OnReceiveYieldProposeSignal(isYield bool) {
	if !isYield {
		return
	}
	if bp.yieldProposing() {
		// halt scheduler execution
		bp.txScheduler.Halt()
		height, _ := bp.ledgerCache.CurrentHeight()
		bp.proposalCache.ResetProposedAt(height + 1)
	}
}

// yieldProposing, to yield proposing handle
func (bp *BlockProposerImpl) yieldProposing() bool {
	// signal finish propose only if proposer is not idle
	bp.idleMu.Lock()
	defer bp.idleMu.Unlock()
	if !bp.idle {
		bp.finishProposeC <- true
		bp.idle = true
		return true
	}
	return false
}

// getDuration, get propose duration from config.
// If not access from config, use default value.
func (bp *BlockProposerImpl) getDuration() time.Duration {
	if bp.chainConf == nil || bp.chainConf.ChainConfig() == nil {
		return DEFAULTDURATION * time.Millisecond
	}
	chainConfig := bp.chainConf.ChainConfig()
	duration := chainConfig.Block.BlockInterval
	if duration <= 0 {
		return DEFAULTDURATION * time.Millisecond
	}
	return time.Duration(duration) * time.Millisecond

}

// getChainVersion, get chain version from config.
// If not access from config, use default value.
// @Deprecated
//nolint: unused
func (bp *BlockProposerImpl) getChainVersion() uint32 {
	//if bp.chainConf == nil || bp.chainConf.ChainConfig() == nil {
	//	return []byte(protocol.DefaultBlockVersion)
	//}
	//return []byte(bp.chainConf.ChainConfig().Version)
	return protocol.DefaultBlockVersion
}

// setNotIdle, set not idle status
func (bp *BlockProposerImpl) setNotIdle() bool {
	bp.idleMu.Lock()
	defer bp.idleMu.Unlock()
	if bp.idle {
		bp.idle = false
		return true
	}
	return false
}

// isIdle, to check if proposer is idle
func (bp *BlockProposerImpl) isIdle() bool {
	bp.idleMu.Lock()
	defer bp.idleMu.Unlock()
	return bp.idle
}

// setIdle, set idle status
func (bp *BlockProposerImpl) setIdle() {
	bp.idleMu.Lock()
	defer bp.idleMu.Unlock()
	bp.idle = true
}

// setIsSelfProposer, set isProposer status of this node
func (bp *BlockProposerImpl) setIsSelfProposer(isSelfProposer bool) {
	bp.proposerMu.Lock()
	defer bp.proposerMu.Unlock()
	bp.isProposer = isSelfProposer
	if !bp.isProposer {
		bp.proposeTimer.Stop()
	} else {
		bp.proposeTimer.Reset(bp.getDuration())
	}
}

// isSelfProposer, return if this node is consensus proposer
func (bp *BlockProposerImpl) isSelfProposer() bool {
	bp.proposerMu.RLock()
	defer bp.proposerMu.RUnlock()
	return bp.isProposer
}

/*
 * shouldProposeByMaxBFT, check if node should propose new block
 * Only for maxbft consensus
 */
func (bp *BlockProposerImpl) shouldProposeByMaxBFT(height uint64, preHash []byte) bool {
	committedBlock := bp.ledgerCache.GetLastCommittedBlock()
	if committedBlock == nil {
		bp.log.Errorf("no committed block found")
		return false
	}
	currentHeight := committedBlock.Header.BlockHeight
	// proposing height must higher than current height
	if currentHeight >= height {
		bp.log.Errorf("current commit block height: %d, propose height: %d", currentHeight, height)
		return false
	}
	if height == currentHeight+1 {
		// height follows the last committed block
		if bytes.Equal(committedBlock.Header.BlockHash, preHash) {
			return true
		}
		bp.log.Errorf("block pre hash error, expect %x, got %x, can not propose",
			committedBlock.Header.BlockHash, preHash)
		return false

	}
	// if height not follows the last committed block, then check last proposed block
	b, _ := bp.proposalCache.GetProposedBlockByHashAndHeight(preHash, height-1)
	if b == nil {
		bp.log.Errorf("not find preBlock: [%d:%x]", height-1, preHash)
	}
	return b != nil
}
