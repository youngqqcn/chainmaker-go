/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"chainmaker.org/chainmaker-go/module/core/common/scheduler"
	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker-go/module/subscriber"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonErrors "chainmaker.org/chainmaker/common/v2/errors"
	"chainmaker.org/chainmaker/common/v2/monitor"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/localconf/v2"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
	batch "chainmaker.org/chainmaker/txpool-batch/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//proposeRepeatTimer *time.Timer //timer controls the propose repeat interval
	//ProposeRepeatTimerMap = make(map[string]*time.Timer)

	ProposeRepeatTimerMap sync.Map
)

const (
	DEFAULTDURATION = 1000 // default proposal duration, millis seconds
	//blockSig:%d,vm:%d,txVerify:%d,txRoot:%d
	BlockSig = "blockSig"
	VM       = "vm"
	TxVerify = "txVerify"
	TxRoot   = "txRoot"
)

type BlockBuilderConf struct {
	ChainId         string                   // chain id, to identity this chain
	TxPool          protocol.TxPool          // tx pool provides tx batch
	TxScheduler     protocol.TxScheduler     // scheduler orders tx batch into DAG form and returns a block
	SnapshotManager protocol.SnapshotManager // snapshot manager
	Identity        protocol.SigningMember   // identity manager
	LedgerCache     protocol.LedgerCache     // ledger cache
	ProposalCache   protocol.ProposalCache
	ChainConf       protocol.ChainConf // chain config
	Log             protocol.Logger
	StoreHelper     conf.StoreHelper
}

type BlockBuilder struct {
	chainId         string                   // chain id, to identity this chain
	txPool          protocol.TxPool          // tx pool provides tx batch
	txScheduler     protocol.TxScheduler     // scheduler orders tx batch into DAG form and returns a block
	snapshotManager protocol.SnapshotManager // snapshot manager
	identity        protocol.SigningMember   // identity manager
	ledgerCache     protocol.LedgerCache     // ledger cache
	proposalCache   protocol.ProposalCache
	chainConf       protocol.ChainConf // chain config
	log             protocol.Logger
	storeHelper     conf.StoreHelper
}

func NewBlockBuilder(conf *BlockBuilderConf) *BlockBuilder {
	creatorBlock := &BlockBuilder{
		chainId:         conf.ChainId,
		txPool:          conf.TxPool,
		txScheduler:     conf.TxScheduler,
		snapshotManager: conf.SnapshotManager,
		identity:        conf.Identity,
		ledgerCache:     conf.LedgerCache,
		proposalCache:   conf.ProposalCache,
		chainConf:       conf.ChainConf,
		log:             conf.Log,
		storeHelper:     conf.StoreHelper,
	}

	return creatorBlock
}

func (bb *BlockBuilder) GenerateNewBlock(proposingHeight uint64, preHash []byte, txBatch []*commonPb.Transaction) (
	*commonPb.Block, []int64, error) {
	timeLasts := make([]int64, 0)
	currentHeight, _ := bb.ledgerCache.CurrentHeight()
	lastBlock := bb.findLastBlockFromCache(proposingHeight, preHash, currentHeight)
	if lastBlock == nil {
		return nil, nil, fmt.Errorf("no pre block found [%d] (%x)", proposingHeight-1, preHash)
	}
	isConfigBlock := false
	if len(txBatch) == 1 && utils.IsConfigTx(txBatch[0]) {
		isConfigBlock = true
	}
	block, err := initNewBlock(lastBlock, bb.identity, bb.chainId, bb.chainConf, isConfigBlock)
	if err != nil {
		return block, timeLasts, err
	}
	if block == nil {
		bb.log.Warnf("generate new block failed, block == nil")
		return nil, timeLasts, fmt.Errorf("generate new block failed, block == nil")
	}
	//if txBatch == nil {
	//	// For ChainedBFT consensus, generate an empty block if tx batch is empty.
	//	return block, timeLasts, nil
	//}

	// validate tx and verify ACL，split into 2 slice according to result
	// validatedTxs are txs passed validate and should be executed by contract
	var aclFailTxs = make([]*commonPb.Transaction, 0) // No need to ACL check, this slice is empty
	var validatedTxs = txBatch

	// txScheduler handle：
	// 1. execute transaction and fill the result, include rw set digest, and remove from txBatch
	// 2. calculate dag and fill into block
	// 3. fill txs into block
	// If only part of the txBatch is filled into the Block, consider executing it again
	ssStartTick := utils.CurrentTimeMillisSeconds()
	snapshot := bb.snapshotManager.NewSnapshot(lastBlock, block)

	beginDbTick := utils.CurrentTimeMillisSeconds()
	bb.storeHelper.BeginDbTransaction(snapshot.GetBlockchainStore(), block.GetTxKey())

	vmStartTick := utils.CurrentTimeMillisSeconds()
	txRWSetMap, contractEventMap, err := bb.txScheduler.Schedule(block, validatedTxs, snapshot)

	ssLasts := beginDbTick - ssStartTick
	dbLasts := vmStartTick - beginDbTick
	vmLasts := utils.CurrentTimeMillisSeconds() - vmStartTick
	timeLasts = append(timeLasts, ssLasts, dbLasts, vmLasts)

	if err != nil {
		return nil, timeLasts, fmt.Errorf("schedule block(%d,%x) error %s",
			block.Header.BlockHeight, block.Header.BlockHash, err)
	}

	// deal with the special situation：
	// 1. only one tx and schedule time out
	// 2. package the empty block
	if !utils.CanProposeEmptyBlock(bb.chainConf.ChainConfig().Consensus.Type) && len(block.Txs) == 0 {
		return nil, timeLasts, fmt.Errorf("no txs in scheduled block, proposing block ends")
	}

	finalizeStartTick := utils.CurrentTimeMillisSeconds()
	err = FinalizeBlock(
		block,
		txRWSetMap,
		aclFailTxs,
		bb.chainConf.ChainConfig().Crypto.Hash,
		bb.log)
	finalizeLasts := utils.CurrentTimeMillisSeconds() - finalizeStartTick
	if err != nil {
		return nil, timeLasts, fmt.Errorf("finalizeBlock block(%d,%s) error %s",
			block.Header.BlockHeight, hex.EncodeToString(block.Header.BlockHash), err)
	}
	timeLasts = append(timeLasts, finalizeLasts)
	// get txs schedule timeout and put back to txpool
	var txsTimeout = make([]*commonPb.Transaction, 0)
	if len(txRWSetMap) < len(txBatch) {
		// if tx not in txRWSetMap, tx should be put back to txpool
		for _, tx := range txBatch {
			if _, ok := txRWSetMap[tx.Payload.TxId]; !ok {
				txsTimeout = append(txsTimeout, tx)
			}
		}
		bb.txPool.RetryAndRemoveTxs(txsTimeout, nil)
	}

	// cache proposed block
	bb.log.Debugf("set proposed block(%d,%x)", block.Header.BlockHeight, block.Header.BlockHash)
	if err = bb.proposalCache.SetProposedBlock(block, txRWSetMap, contractEventMap, true); err != nil {
		return block, timeLasts, err
	}
	bb.proposalCache.SetProposedAt(block.Header.BlockHeight)

	return block, timeLasts, nil
}

func (bb *BlockBuilder) findLastBlockFromCache(proposingHeight uint64, preHash []byte,
	currentHeight uint64) *commonPb.Block {
	var lastBlock *commonPb.Block
	if currentHeight+1 == proposingHeight {
		lastBlock = bb.ledgerCache.GetLastCommittedBlock()
	} else {
		lastBlock, _ = bb.proposalCache.GetProposedBlockByHashAndHeight(preHash, proposingHeight-1)
	}
	return lastBlock
}

func initNewBlock(
	lastBlock *commonPb.Block,
	identity protocol.SigningMember,
	chainId string,
	chainConf protocol.ChainConf, isConfigBlock bool) (*commonPb.Block, error) {
	// get node pk from identity
	proposer, err := identity.GetMember()
	if err != nil {
		return nil, fmt.Errorf("identity serialize failed, %s", err)
	}
	preConfHeight := lastBlock.Header.PreConfHeight
	// if last block is config block, then this block.preConfHeight is last block height
	if utils.IsConfBlock(lastBlock) {
		preConfHeight = lastBlock.Header.BlockHeight
	}

	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:        chainId,
			BlockHeight:    lastBlock.Header.BlockHeight + 1,
			PreBlockHash:   lastBlock.Header.BlockHash,
			BlockHash:      nil,
			PreConfHeight:  preConfHeight,
			BlockVersion:   protocol.DefaultBlockVersion,
			DagHash:        nil,
			RwSetRoot:      nil,
			TxRoot:         nil,
			BlockTimestamp: utils.CurrentTimeSeconds(),
			Proposer:       proposer,
			ConsensusArgs:  nil,
			TxCount:        0,
			Signature:      nil,
		},
		Dag:            &commonPb.DAG{},
		Txs:            nil,
		AdditionalData: nil,
	}
	if isConfigBlock {
		block.Header.BlockType = commonPb.BlockType_CONFIG_BLOCK
	}
	return block, nil
}

func FinalizeBlock(
	block *commonPb.Block,
	txRWSetMap map[string]*commonPb.TxRWSet,
	aclFailTxs []*commonPb.Transaction,
	hashType string,
	logger protocol.Logger) error {

	if aclFailTxs != nil && len(aclFailTxs) > 0 { //nolint: gosimple
		// append acl check failed txs to the end of block.Txs
		block.Txs = append(block.Txs, aclFailTxs...)
	}

	// TxCount contains acl verify failed txs and invoked contract txs
	txCount := len(block.Txs)
	block.Header.TxCount = uint32(txCount)

	// TxRoot/RwSetRoot
	var errs []error
	txHashes := make([][]byte, txCount)
	wg := &sync.WaitGroup{}
	wg.Add(txCount)
	for i, tx := range block.Txs {
		// finalize tx, put rwsethash into tx.Result
		rwSet := txRWSetMap[tx.Payload.TxId]
		if rwSet == nil {
			rwSet = &commonPb.TxRWSet{
				TxId:     tx.Payload.TxId,
				TxReads:  nil,
				TxWrites: nil,
			}
		}
		go func(tx *commonPb.Transaction, rwSet *commonPb.TxRWSet, x int) {
			defer wg.Done()
			var err error
			txHashes[x], err = getTxHash(tx, rwSet, hashType, logger)
			if err != nil {
				errs = append(errs, err)
			}

		}(tx, rwSet, i)
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	wg.Add(3)
	//calc tx root
	go func() {
		defer wg.Done()
		var err error
		block.Header.TxRoot, err = hash.GetMerkleRoot(hashType, txHashes)
		if err != nil {
			logger.Warnf("get tx merkle root error %s", err)
			errs = append(errs, err)
		}
		logger.DebugDynamic(func() string {
			return fmt.Sprintf("GetMerkleRoot(%s) get %x", hashType, block.Header.TxRoot)
		})
	}()
	//calc rwset root
	go func() {
		defer wg.Done()
		var err error
		block.Header.RwSetRoot, err = utils.CalcRWSetRoot(hashType, block.Txs)
		if err != nil {
			logger.Warnf("get rwset merkle root error %s", err)
			errs = append(errs, err)
		}
	}()
	//calc dag hash
	go func() {
		defer wg.Done()
		// DagDigest
		var dagHash []byte
		var err error
		dagHash, err = utils.CalcDagHash(hashType, block.Dag)
		if err != nil {
			logger.Warnf("get dag hash error %s", err)
			errs = append(errs, err)
		}
		block.Header.DagHash = dagHash
	}()
	wg.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
func getTxHash(tx *commonPb.Transaction, rwSet *commonPb.TxRWSet, hashType string, logger protocol.Logger) (
	[]byte, error) {
	var rwSetHash []byte
	rwSetHash, err := utils.CalcRWSetHash(hashType, rwSet)
	logger.DebugDynamic(func() string {
		str := fmt.Sprintf("CalcRWSetHash rwset: %+v ,hash: %x", rwSet, rwSetHash)
		if len(str) > 1024 {
			str = str[:1024] + " ......"
		}
		return str
	})
	if err != nil {
		return nil, err
	}
	if tx.Result == nil {
		// in case tx.Result is nil, avoid panic
		e := fmt.Errorf("tx(%s) result == nil", tx.Payload.TxId)
		logger.Error(e.Error())
		return nil, e
	}
	tx.Result.RwSetHash = rwSetHash
	// calculate complete tx hash, include tx.Header, tx.Payload, tx.Result
	var txHash []byte
	txHash, err = utils.CalcTxHash(hashType, tx)
	if err != nil {
		return nil, err
	}
	return txHash, nil
}

// IsTxCountValid, to check if txcount in block is valid
func IsTxCountValid(block *commonPb.Block) error {
	if block.Header.TxCount != uint32(len(block.Txs)) {
		return fmt.Errorf("txcount expect %d, got %d", block.Header.TxCount, len(block.Txs))
	}
	return nil
}

// IsHeightValid, to check if block height is valid
func IsHeightValid(block *commonPb.Block, currentHeight uint64) error {
	if currentHeight+1 != block.Header.BlockHeight {
		return fmt.Errorf("height expect %d, got %d", currentHeight+1, block.Header.BlockHeight)
	}
	return nil
}

// IsPreHashValid, to check if block.preHash equals with last block hash
func IsPreHashValid(block *commonPb.Block, preHash []byte) error {
	if !bytes.Equal(preHash, block.Header.PreBlockHash) {
		return fmt.Errorf("prehash expect %x, got %x", preHash, block.Header.PreBlockHash)
	}
	return nil
}

// IsBlockHashValid, to check if block hash equals with result calculated from block
func IsBlockHashValid(block *commonPb.Block, hashType string) error {
	hash, err := utils.CalcBlockHash(hashType, block)
	if err != nil {
		return fmt.Errorf("calc block hash error")
	}
	if !bytes.Equal(hash, block.Header.BlockHash) {
		return fmt.Errorf("block hash expect %x, got %x", block.Header.BlockHash, hash)
	}
	return nil
}

// IsTxDuplicate, to check if there is duplicated transactions in one block
func IsTxDuplicate(txs []*commonPb.Transaction) bool {
	txSet := make(map[string]struct{})
	exist := struct{}{}
	for _, tx := range txs {
		if tx == nil || tx.Payload == nil {
			return true
		}
		txSet[tx.Payload.TxId] = exist
	}
	// length of set < length of txs, means txs have duplicate tx
	return len(txSet) < len(txs)
}

// IsMerkleRootValid, to check if block merkle root equals with simulated merkle root
func IsMerkleRootValid(block *commonPb.Block, txHashes [][]byte, hashType string) error {
	txRoot, err := hash.GetMerkleRoot(hashType, txHashes)
	if err != nil || !bytes.Equal(txRoot, block.Header.TxRoot) {
		return fmt.Errorf("GetMerkleRoot(%s,%v) get %x ,txroot expect %x, got %x, err: %s",
			hashType, txHashes, txRoot, block.Header.TxRoot, txRoot, err)
	}
	return nil
}

// IsDagHashValid, to check if block dag equals with simulated block dag
func IsDagHashValid(block *commonPb.Block, hashType string) error {
	dagHash, err := utils.CalcDagHash(hashType, block.Dag)
	if err != nil || !bytes.Equal(dagHash, block.Header.DagHash) {
		return fmt.Errorf("dag expect %x, got %x", block.Header.DagHash, dagHash)
	}
	return nil
}

// IsRWSetHashValid, to check if read write set is valid
func IsRWSetHashValid(block *commonPb.Block, hashType string) error {
	rwSetRoot, err := utils.CalcRWSetRoot(hashType, block.Txs)
	if err != nil {
		return fmt.Errorf("calc rwset error, %s", err)
	}
	if !bytes.Equal(rwSetRoot, block.Header.RwSetRoot) {
		return fmt.Errorf("rwset expect %x, got %x", block.Header.RwSetRoot, rwSetRoot)
	}
	return nil
}

// getChainVersion, get chain version from config.
// If not access from config, use default value.
// @Deprecated
//func getChainVersion(chainConf protocol.ChainConf) []byte {
//	if chainConf == nil || chainConf.ChainConfig() == nil {
//		return []byte(protocol.DefaultBlockVersion)
//	}
//	return []byte(chainConf.ChainConfig().Version)
//}

func VerifyHeight(height uint64, ledgerCache protocol.LedgerCache) error {
	currentHeight, err := ledgerCache.CurrentHeight()
	if err != nil {
		return err
	}
	if currentHeight+1 != height {
		return fmt.Errorf("verify height fail,expected [%d]", currentHeight+1)
	}
	return nil
}

func CheckBlockDigests(block *commonPb.Block, txHashes [][]byte, hashType string, log protocol.Logger) error {
	if err := IsMerkleRootValid(block, txHashes, hashType); err != nil {
		log.Error(err)
		return err
	}
	// verify DAG hash
	if err := IsDagHashValid(block, hashType); err != nil {
		log.Error(err)
		return err
	}
	// verify read write set, check if simulate result is equal with rwset in block header
	if err := IsRWSetHashValid(block, hashType); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func CheckVacuumBlock(block *commonPb.Block, consensusType consensus.ConsensusType) error {
	if block.Header.TxCount == 0 {
		if utils.CanProposeEmptyBlock(consensusType) {
			// for consensus that allows empty block, skip txs verify
			return nil
		}

		// for consensus that NOT allows empty block, return error
		return fmt.Errorf("tx must not empty")
	}
	return nil
}

type VerifierBlockConf struct {
	ChainConf       protocol.ChainConf
	Log             protocol.Logger
	LedgerCache     protocol.LedgerCache
	Ac              protocol.AccessControlProvider
	SnapshotManager protocol.SnapshotManager
	VmMgr           protocol.VmManager
	TxPool          protocol.TxPool
	BlockchainStore protocol.BlockchainStore
	ProposalCache   protocol.ProposalCache // proposal cache
	StoreHelper     conf.StoreHelper
	TxScheduler     protocol.TxScheduler
}

type VerifierBlock struct {
	chainConf       protocol.ChainConf
	log             protocol.Logger
	ledgerCache     protocol.LedgerCache
	ac              protocol.AccessControlProvider
	snapshotManager protocol.SnapshotManager
	vmMgr           protocol.VmManager
	txScheduler     protocol.TxScheduler
	txPool          protocol.TxPool
	blockchainStore protocol.BlockchainStore
	proposalCache   protocol.ProposalCache // proposal cache
	storeHelper     conf.StoreHelper
}

func NewVerifierBlock(conf *VerifierBlockConf) *VerifierBlock {
	verifyBlock := &VerifierBlock{
		chainConf:       conf.ChainConf,
		log:             conf.Log,
		ledgerCache:     conf.LedgerCache,
		ac:              conf.Ac,
		snapshotManager: conf.SnapshotManager,
		vmMgr:           conf.VmMgr,
		txPool:          conf.TxPool,
		blockchainStore: conf.BlockchainStore,
		proposalCache:   conf.ProposalCache,
		storeHelper:     conf.StoreHelper,
		txScheduler:     conf.TxScheduler,
	}
	var schedulerFactory scheduler.TxSchedulerFactory
	verifyBlock.txScheduler = schedulerFactory.NewTxScheduler(
		verifyBlock.vmMgr,
		verifyBlock.chainConf,
		conf.StoreHelper,
	)
	return verifyBlock
}

func (vb *VerifierBlock) FetchLastBlock(block *commonPb.Block) (*commonPb.Block, error) { //nolint: staticcheck
	currentHeight, _ := vb.ledgerCache.CurrentHeight()
	if currentHeight >= block.Header.BlockHeight {
		return nil, commonErrors.ErrBlockHadBeenCommited
	}

	var lastBlock *commonPb.Block
	if currentHeight+1 == block.Header.BlockHeight {
		lastBlock = vb.ledgerCache.GetLastCommittedBlock() //nolint: staticcheck
	} else {
		lastBlock, _ = vb.proposalCache.GetProposedBlockByHashAndHeight(
			block.Header.PreBlockHash, block.Header.BlockHeight-1)
	}
	if lastBlock == nil {
		return nil, fmt.Errorf("no pre block found [%d](%x)", block.Header.BlockHeight-1, block.Header.PreBlockHash)
	}
	return lastBlock, nil
}

// validateBlock, validate block and transactions
func (vb *VerifierBlock) ValidateBlock(
	block, lastBlock *commonPb.Block, hashType string, timeLasts map[string]int64) (
	map[string]*commonPb.TxRWSet, map[string][]*commonPb.ContractEvent, map[string]int64, error) {

	if err := IsBlockHashValid(block, vb.chainConf.ChainConfig().Crypto.Hash); err != nil {
		return nil, nil, timeLasts, err
	}

	// verify block sig and also verify identity and auth of block proposer
	startSigTick := utils.CurrentTimeMillisSeconds()
	vb.log.DebugDynamic(func() string {
		return fmt.Sprintf("verify block \n %s", utils.FormatBlock(block))
	})
	if ok, err := utils.VerifyBlockSig(hashType, block, vb.ac); !ok || err != nil {
		return nil, nil, timeLasts, fmt.Errorf("(%d,%x - %x,%x) [signature]",
			block.Header.BlockHeight, block.Header.BlockHash, block.Header.Proposer, block.Header.Signature)
	}
	sigLasts := utils.CurrentTimeMillisSeconds() - startSigTick
	timeLasts[BlockSig] = sigLasts

	err := CheckVacuumBlock(block, vb.chainConf.ChainConfig().Consensus.Type)
	if err != nil {
		return nil, nil, timeLasts, err
	}
	// we must new a snapshot for the vacant block,
	// otherwise the subsequent snapshot can not link to the previous snapshot.
	snapshotTick := utils.CurrentTimeMillisSeconds()
	snapshot := vb.snapshotManager.NewSnapshot(lastBlock, block)
	if len(block.Txs) == 0 {
		return nil, nil, timeLasts, nil
	}
	// verify if txs are duplicate in this block
	if IsTxDuplicate(block.Txs) {
		return nil, nil, timeLasts, fmt.Errorf("tx duplicate")
	}

	// simulate with DAG, and verify read write set
	startDbTxTick := utils.CurrentTimeMillisSeconds()
	vb.storeHelper.BeginDbTransaction(snapshot.GetBlockchainStore(), block.GetTxKey())

	startVMTick := utils.CurrentTimeMillisSeconds()
	txRWSetMap, txResultMap, err := vb.txScheduler.SimulateWithDag(block, snapshot)
	vmLasts := utils.CurrentTimeMillisSeconds() - startVMTick
	vb.log.Infof("Validate block[%v](txs:%v), time used(new snapshot:%v, start DB transaction:%v, vm:%v)",
		block.Header.BlockHeight, block.Header.TxCount, startDbTxTick-snapshotTick, startVMTick-startDbTxTick, vmLasts)

	timeLasts[VM] = vmLasts
	if err != nil {
		return nil, nil, timeLasts, fmt.Errorf("simulate %s", err)
	}
	if block.Header.TxCount != uint32(len(txRWSetMap)) {
		return nil, nil, timeLasts, fmt.Errorf("simulate txcount expect %d, got %d",
			block.Header.TxCount, len(txRWSetMap))
	}

	// 2.transaction verify
	startTxTick := utils.CurrentTimeMillisSeconds()
	verifierTxConf := &VerifierTxConfig{
		Block:       block,
		TxResultMap: txResultMap,
		TxRWSetMap:  txRWSetMap,
		ChainConf:   vb.chainConf,
		Log:         vb.log,
		Ac:          vb.ac,
		TxPool:      vb.txPool,
		Store:       vb.blockchainStore,
	}
	verifiertx := NewVerifierTx(verifierTxConf)
	txHashes, _, errTxs, err := verifiertx.verifierTxs(block)
	txLasts := utils.CurrentTimeMillisSeconds() - startTxTick
	timeLasts[TxVerify] = txLasts
	if err != nil {
		if len(errTxs) > 0 {
			vb.log.Warn("[Duplicate txs] delete the err txs")
			vb.txPool.RetryAndRemoveTxs(nil, errTxs)
		}
		return nil, nil, timeLasts, fmt.Errorf("verify failed [%d](%x), %s ",
			block.Header.BlockHeight, block.Header.BlockHash, err)
	}
	//if protocol.CONSENSUS_VERIFY == mode && len(newAddTx) > 0 {
	//	v.txPool.AddTrustedTx(newAddTx)
	//}

	// get contract events
	contractEventMap := make(map[string][]*commonPb.ContractEvent)
	for _, tx := range block.Txs {
		var events []*commonPb.ContractEvent
		if result, ok := txResultMap[tx.Payload.TxId]; ok {
			events = result.ContractResult.ContractEvent
		}
		contractEventMap[tx.Payload.TxId] = events
	}
	// verify TxRoot
	startRootsTick := utils.CurrentTimeMillisSeconds()
	err = CheckBlockDigests(block, txHashes, hashType, vb.log)
	if err != nil {
		return txRWSetMap, contractEventMap, timeLasts, err
	}
	rootsLast := utils.CurrentTimeMillisSeconds() - startRootsTick
	timeLasts[TxRoot] = rootsLast

	return txRWSetMap, contractEventMap, timeLasts, nil
}

// validateBlock, validate block and transactions
func (vb *VerifierBlock) ValidateBlockWithRWSets(
	block, lastBlock *commonPb.Block, hashType string,
	timeLasts map[string]int64, txRWSetMap map[string]*commonPb.TxRWSet) (
	map[string][]*commonPb.ContractEvent, map[string]int64, error) {
	// 1.block verify
	if err := IsBlockHashValid(block, vb.chainConf.ChainConfig().Crypto.Hash); err != nil {
		return nil, timeLasts, err
	}
	txResultMap := make(map[string]*commonPb.Result)
	for _, tx := range block.GetTxs() {
		if tx.Result != nil {
			txResultMap[tx.Payload.TxId] = tx.Result
		}
	}
	// verify block sig and also verify identity and auth of block proposer
	startSigTick := utils.CurrentTimeMillisSeconds()
	vb.log.DebugDynamic(func() string {
		return fmt.Sprintf("verify block \n %s", utils.FormatBlock(block))
	})
	if ok, err := utils.VerifyBlockSig(hashType, block, vb.ac); !ok || err != nil {
		return nil, timeLasts, fmt.Errorf("(%d,%x - %x,%x) [signature]",
			block.Header.BlockHeight, block.Header.BlockHash, block.Header.Proposer, block.Header.Signature)
	}
	sigLasts := utils.CurrentTimeMillisSeconds() - startSigTick
	timeLasts[BlockSig] = sigLasts

	err := CheckVacuumBlock(block, vb.chainConf.ChainConfig().Consensus.Type)
	if err != nil {
		return nil, timeLasts, err
	}
	// we must new a snapshot for the vacant block,
	// otherwise the subsequent snapshot can not link to the previous snapshot.
	snapshot := vb.snapshotManager.NewSnapshot(lastBlock, block)
	if len(block.Txs) == 0 {
		return nil, timeLasts, nil
	}
	// verify if txs are duplicate in this block
	if IsTxDuplicate(block.Txs) {
		return nil, timeLasts, fmt.Errorf("tx duplicate")
	}

	// simulate with DAG, and verify read write set
	startVMTick := utils.CurrentTimeMillisSeconds()
	vb.storeHelper.BeginDbTransaction(snapshot.GetBlockchainStore(), block.GetTxKey())
	//txRWSetMap, txResultMap, err := vb.txScheduler.SimulateWithDag(block, snapshot)
	//if err != nil {
	//	return nil, nil, timeLasts, fmt.Errorf("simulate %s", err)
	//}

	vmLasts := utils.CurrentTimeMillisSeconds() - startVMTick
	timeLasts[VM] = vmLasts

	if block.Header.TxCount != uint32(len(txRWSetMap)) {
		return nil, timeLasts, fmt.Errorf("simulate txcount expect %d, got %d",
			block.Header.TxCount, len(txRWSetMap))
	}

	// 2.transaction verify
	startTxTick := utils.CurrentTimeMillisSeconds()
	verifierTxConf := &VerifierTxConfig{
		Block:       block,
		TxResultMap: txResultMap,
		TxRWSetMap:  txRWSetMap,
		ChainConf:   vb.chainConf,
		Log:         vb.log,
		Ac:          vb.ac,
		TxPool:      vb.txPool,
		Store:       vb.blockchainStore,
	}
	verifiertx := NewVerifierTx(verifierTxConf)
	txHashes, _, errTxs, err := verifiertx.verifierTxs(block)
	vb.log.Warnf("verifierTxs txHashCount:%d, txCount:%d, %x", len(txHashes), len(block.Txs), block.Header.TxRoot)
	txLasts := utils.CurrentTimeMillisSeconds() - startTxTick
	timeLasts[TxVerify] = txLasts
	if err != nil {
		if len(errTxs) > 0 {
			vb.log.Warn("[Duplicate txs] delete the err txs")
			vb.txPool.RetryAndRemoveTxs(nil, errTxs)
		}
		return nil, timeLasts, fmt.Errorf("verify failed [%d](%x), %s ",
			block.Header.BlockHeight, block.Header.BlockHash, err)
	}
	//if protocol.CONSENSUS_VERIFY == mode && len(newAddTx) > 0 {
	//	v.txPool.AddTrustedTx(newAddTx)
	//}

	// get contract events
	contractEventMap := make(map[string][]*commonPb.ContractEvent)
	for _, tx := range block.Txs {
		var events []*commonPb.ContractEvent
		if result, ok := txResultMap[tx.Payload.TxId]; ok {
			events = result.ContractResult.ContractEvent
		}
		contractEventMap[tx.Payload.TxId] = events
	}
	// verify TxRoot
	startRootsTick := utils.CurrentTimeMillisSeconds()
	err = CheckBlockDigests(block, txHashes, hashType, vb.log)
	if err != nil {
		return contractEventMap, timeLasts, err
	}
	rootsLast := utils.CurrentTimeMillisSeconds() - startRootsTick
	timeLasts[TxRoot] = rootsLast

	return contractEventMap, timeLasts, nil
}

//nolint: staticcheck
func CheckPreBlock(block *commonPb.Block, lastBlock *commonPb.Block,
	err error, lastBlockHash []byte, proposedHeight uint64) error {

	if err = IsHeightValid(block, proposedHeight); err != nil {
		return err
	}
	// check if this block pre hash is equal with last block hash
	return IsPreHashValid(block, lastBlockHash)
}

// BlockCommitterImpl implements BlockCommitter interface.
// To commit a block after it is confirmed by consensus module.
type BlockCommitterImpl struct {
	chainId string // chain id, to identity this chain
	// Store is a block store that will only fetch data locally
	blockchainStore protocol.BlockchainStore // blockchain store
	snapshotManager protocol.SnapshotManager // snapshot manager
	txPool          protocol.TxPool          // transaction pool
	chainConf       protocol.ChainConf       // chain config

	ledgerCache             protocol.LedgerCache        // ledger cache
	proposalCache           protocol.ProposalCache      // proposal cache
	log                     protocol.Logger             // logger
	msgBus                  msgbus.MessageBus           // message bus
	mu                      sync.Mutex                  // lock, to avoid concurrent block commit
	subscriber              *subscriber.EventSubscriber // subscriber
	verifier                protocol.BlockVerifier      // block verifier
	commonCommit            *CommitBlock
	metricBlockSize         *prometheus.HistogramVec // metric block size
	metricBlockCounter      *prometheus.CounterVec   // metric block counter
	metricTxCounter         *prometheus.CounterVec   // metric transaction counter
	metricBlockCommitTime   *prometheus.HistogramVec // metric block commit time
	metricBlockIntervalTime *prometheus.HistogramVec // metric block interval time
	metricTpsGauge          *prometheus.GaugeVec     // metric real-time transaction per second (TPS)
	storeHelper             conf.StoreHelper
	blockInterval           int64
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
	StoreHelper     conf.StoreHelper
}

func NewBlockCommitter(config BlockCommitterConfig, log protocol.Logger) (protocol.BlockCommitter, error) {
	blockchain := &BlockCommitterImpl{
		chainId:         config.ChainId,
		blockchainStore: config.BlockchainStore,
		snapshotManager: config.SnapshotManager,
		txPool:          config.TxPool,
		ledgerCache:     config.LedgerCache,
		proposalCache:   config.ProposedCache,
		log:             log,
		chainConf:       config.ChainConf,
		msgBus:          config.MsgBus,
		subscriber:      config.Subscriber,
		verifier:        config.Verifier,
		storeHelper:     config.StoreHelper,
	}

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		blockchain.metricBlockSize = monitor.NewHistogramVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricBlockSize,
			monitor.HelpCurrentBlockSizeMetric,
			prometheus.ExponentialBuckets(1024, 2, 12),
			monitor.ChainId,
		)

		blockchain.metricBlockCounter = monitor.NewCounterVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricBlockCounter,
			monitor.HelpBlockCountsMetric,
			monitor.ChainId,
		)

		blockchain.metricTxCounter = monitor.NewCounterVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricTxCounter,
			monitor.HelpTxCountsMetric,
			monitor.ChainId,
		)

		blockchain.metricBlockCommitTime = monitor.NewHistogramVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricBlockCommitTime,
			monitor.HelpBlockCommitTimeMetric,
			[]float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 10},
			monitor.ChainId,
		)

		blockchain.metricBlockIntervalTime = monitor.NewHistogramVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricBlockIntervalTime,
			monitor.HelpBlockIntervalTimeMetric,
			[]float64{0.2, 0.5, 1, 2, 5, 10, 20},
			monitor.ChainId,
		)

		blockchain.metricTpsGauge = monitor.NewGaugeVec(
			monitor.SUBSYSTEM_CORE_COMMITTER,
			monitor.MetricTpsGauge,
			monitor.HelpTpsGaugeMetric,
			monitor.ChainId,
		)
	}

	cbConf := &CommitBlockConf{
		Store:                   blockchain.blockchainStore,
		Log:                     blockchain.log,
		SnapshotManager:         blockchain.snapshotManager,
		TxPool:                  blockchain.txPool,
		LedgerCache:             blockchain.ledgerCache,
		ChainConf:               blockchain.chainConf,
		MsgBus:                  blockchain.msgBus,
		MetricBlockCommitTime:   blockchain.metricBlockCommitTime,
		MetricBlockIntervalTime: blockchain.metricBlockIntervalTime,
		MetricBlockCounter:      blockchain.metricBlockCounter,
		MetricBlockSize:         blockchain.metricBlockSize,
		MetricTxCounter:         blockchain.metricTxCounter,
		MetricTpsGauge:          blockchain.metricTpsGauge,
	}
	blockchain.commonCommit = NewCommitBlock(cbConf)

	return blockchain, nil
}

func (chain *BlockCommitterImpl) isBlockLegal(blk *commonPb.Block) error {
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

func (chain *BlockCommitterImpl) AddBlock(block *commonPb.Block) (err error) {
	defer func() {
		panicErr := recover()
		if err == nil {
			if panicErr != nil {
				err = fmt.Errorf(fmt.Sprint(panicErr))
			} else {
				return
			}
		}
		// rollback sql
		if err == commonErrors.ErrBlockHadBeenCommited {
			chain.log.Warn("cache add block err: ", err)
		} else {
			chain.log.Error("cache add block err: ", err)
		}

		if sqlErr := chain.storeHelper.RollBack(block, chain.blockchainStore); sqlErr != nil {
			chain.log.Errorf("block [%d] rollback sql failed: %s", block.Header.BlockHeight, sqlErr)
		}
	}()

	startTick := utils.CurrentTimeMillisSeconds()
	chain.log.Debugf("add block(%d,%x)=(%x,%d,%d)",
		block.Header.BlockHeight, block.Header.BlockHash, block.Header.PreBlockHash,
		block.Header.TxCount, len(block.Txs))
	chain.mu.Lock()
	defer chain.mu.Unlock()

	height := block.Header.BlockHeight
	if err = chain.isBlockLegal(block); err != nil {
		if err == commonErrors.ErrBlockHadBeenCommited {
			chain.log.Warnf("block illegal [%d](hash:%x), %s", height, block.Header.BlockHash, err)
			return err
		}

		chain.log.Errorf("block illegal [%d](hash:%x), %s", height, block.Header.BlockHash, err)
		return err
	}
	lastProposed, rwSetMap, conEventMap := chain.proposalCache.GetProposedBlock(block)
	if lastProposed == nil {
		if lastProposed, rwSetMap, conEventMap, err = chain.checkLastProposedBlock(block); err != nil {
			return err
		}
	} else if IfOpenConsensusMessageTurbo(chain.chainConf) {
		// recover the block for proposer when enable the conensus message turbo function.
		lastProposed.Header = block.Header
	}

	// put consensus qc into block
	lastProposed.AdditionalData = block.AdditionalData

	checkLasts := utils.CurrentTimeMillisSeconds() - startTick
	dbLasts, snapshotLasts, confLasts, otherLasts, pubEvent, blockInfo, err := chain.commonCommit.CommitBlock(
		lastProposed, rwSetMap, conEventMap)
	if err != nil {
		chain.log.Errorf("block common commit failed: %s, blockHeight: (%d)",
			err.Error(), lastProposed.Header.BlockHeight)
	}

	// Remove txs from txpool. Remove will invoke proposeSignal from txpool if pool size > txcount
	startPoolTick := utils.CurrentTimeMillisSeconds()
	txRetry := chain.syncWithTxPool(lastProposed, height)
	chain.log.Infof("remove txs[%d] and retry txs[%d] in add block", len(lastProposed.Txs), len(txRetry))
	chain.txPool.RetryAndRemoveTxs(txRetry, lastProposed.Txs)
	poolLasts := utils.CurrentTimeMillisSeconds() - startPoolTick

	chain.proposalCache.ClearProposedBlockAt(height)

	// clear propose repeat map before send
	ProposeRepeatTimerMap = sync.Map{}

	// synchronize new block height to consensus and sync module
	chain.msgBus.PublishSafe(msgbus.BlockInfo, blockInfo)

	curTime := utils.CurrentTimeMillisSeconds()
	elapsed := curTime - startTick
	interval := curTime - chain.blockInterval
	chain.blockInterval = curTime
	chain.log.Infof(
		"commit block [%d](count:%d,hash:%x)"+
			"time used(check:%d,db:%d,ss:%d,conf:%d,pool:%d,pubConEvent:%d,other:%d,total:%d,interval:%d)",
		height, lastProposed.Header.TxCount, lastProposed.Header.BlockHash,
		checkLasts, dbLasts, snapshotLasts, confLasts, poolLasts, pubEvent, otherLasts, elapsed, interval)
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		chain.metricBlockCommitTime.WithLabelValues(chain.chainId).Observe(float64(elapsed) / 1000)
		chain.metricBlockIntervalTime.WithLabelValues(chain.chainId).Observe(float64(interval) / 1000)
		chain.metricTpsGauge.WithLabelValues(chain.chainId).
			Set(float64(lastProposed.Header.TxCount) / (float64(interval) / 1000))
	}
	return nil
}

func (chain *BlockCommitterImpl) syncWithTxPool(block *commonPb.Block, height uint64) []*commonPb.Transaction {
	proposedBlocks := chain.proposalCache.GetProposedBlocksAt(height)
	txRetry := make([]*commonPb.Transaction, 0, len(block.Txs))
	chain.log.Debugf("has %d blocks in height: %d", len(proposedBlocks), height)
	keepTxs := make(map[string]struct{}, len(block.Txs))
	for _, tx := range block.Txs {
		keepTxs[tx.Payload.TxId] = struct{}{}
	}
	for _, b := range proposedBlocks {
		if bytes.Equal(b.Header.BlockHash, block.Header.BlockHash) {
			continue
		}
		for _, tx := range b.Txs {
			if _, ok := keepTxs[tx.Payload.TxId]; !ok {
				txRetry = append(txRetry, tx)
			}
		}
	}
	return txRetry
}

//nolint: ineffassign, staticcheck
func (chain *BlockCommitterImpl) checkLastProposedBlock(block *commonPb.Block) (
	*commonPb.Block, map[string]*commonPb.TxRWSet, map[string][]*commonPb.ContractEvent, error) {
	err := chain.verifier.VerifyBlock(block, protocol.SYNC_VERIFY)
	if err != nil {
		chain.log.Error("block verify failed [%d](hash:%x), %s",
			block.Header.BlockHeight, block.Header.BlockHash, err)
		return nil, nil, nil, err
	}

	lastProposed, rwSetMap, conEventMap := chain.proposalCache.GetProposedBlock(block)
	if lastProposed == nil {
		chain.log.Error("block not verified [%d](hash:%x)", block.Header.BlockHeight, block.Header.BlockHash)
		return lastProposed, rwSetMap, conEventMap,
			fmt.Errorf("block not verified [%d](hash:%x)", block.Header.BlockHeight, block.Header.BlockHash)
	}
	return lastProposed, rwSetMap, conEventMap, nil
}

func IfOpenConsensusMessageTurbo(chainConf protocol.ChainConf) bool {
	value, ok := localconf.ChainMakerConfig.TxPoolConfig["pool_type"]
	if ok {
		txPoolType, _ := value.(string)
		txPoolType = strings.ToUpper(txPoolType)

		if chainConf.ChainConfig().Core.ConsensusTurboConfig.ConsensusMessageTurbo && txPoolType == batch.TxPoolType {
			return true
		}
	}

	return false
}

func RecoverBlock(
	block *commonPb.Block,
	mode protocol.VerifyMode,
	chainConf protocol.ChainConf,
	txPool protocol.TxPool, logger protocol.Logger) (*commonPb.Block, error) {

	if IfOpenConsensusMessageTurbo(chainConf) && protocol.SYNC_VERIFY != mode {
		newBlock := &commonPb.Block{
			Header:         block.Header,
			Dag:            block.Dag,
			Txs:            make([]*commonPb.Transaction, len(block.Txs)),
			AdditionalData: block.AdditionalData,
		}

		txIds := utils.GetTxIds(block.Txs)
		txsMap := make(map[string]*commonPb.Transaction)
		maxRetryTime := chainConf.ChainConfig().Core.ConsensusTurboConfig.RetryTime
		retryInterval := chainConf.ChainConfig().Core.ConsensusTurboConfig.RetryInterval
		for i := uint64(0); i < maxRetryTime; i++ {
			txsMap, _ = txPool.GetTxsByTxIds(txIds)
			if len(txsMap) == len(block.Txs) {
				break
			}
			logger.Debugf("txs map is not map with tx count,height[%d],map[%d],txcount[%d],retry[%d]",
				block.Header.BlockHeight, len(txsMap), block.Header.TxCount, i+1)
			if i+1 == maxRetryTime {
				logger.Debugf("get txs by branchId fail,height[%d],map[%d],txcount[%d]",
					block.Header.BlockHeight, len(txsMap), block.Header.TxCount)
				return nil, fmt.Errorf("block[%d] verify time out error", block.Header.BlockHeight)
			}
			time.Sleep(time.Millisecond * time.Duration(retryInterval))
		}

		for i := range block.Txs {
			newBlock.Txs[i] = txsMap[block.Txs[i].Payload.TxId]
			newBlock.Txs[i].Result = block.Txs[i].Result
			logger.Debugf("recover the block[%d], TxId[%s, %s]",
				newBlock.Header.BlockHeight, newBlock.Txs[i].Payload.TxId, newBlock.Txs[i].Payload.ContractName)
		}

		return newBlock, nil
	}

	// new a block to avoid use the same pointer with consensus.
	return &commonPb.Block{
		Header:         block.Header,
		Dag:            block.Dag,
		Txs:            block.Txs,
		AdditionalData: block.AdditionalData,
	}, nil
}
