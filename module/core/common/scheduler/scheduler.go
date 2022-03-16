/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/hokaccha/go-prettyjson"

	"chainmaker.org/chainmaker/localconf/v2"

	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"chainmaker.org/chainmaker/vm-native/v2/accountmgr"
	"chainmaker.org/chainmaker/vm/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ScheduleTimeout        = 10
	ScheduleWithDagTimeout = 20
)

// TxScheduler transaction scheduler structure
type TxScheduler struct {
	lock            sync.Mutex
	VmManager       protocol.VmManager
	scheduleFinishC chan bool
	log             protocol.Logger
	chainConf       protocol.ChainConf // chain config

	metricVMRunTime *prometheus.HistogramVec
	StoreHelper     conf.StoreHelper
	keyReg          *regexp.Regexp
}

// Transaction dependency in adjacency table representation
type dagNeighbors map[int]bool

// Schedule according to a batch of transactions, and generating DAG according to the conflict relationship
func (ts *TxScheduler) Schedule(block *commonPb.Block, txBatch []*commonPb.Transaction,
	snapshot protocol.Snapshot) (map[string]*commonPb.TxRWSet, map[string][]*commonPb.ContractEvent, error) {

	ts.lock.Lock()
	defer ts.lock.Unlock()
	txBatchSize := len(txBatch)
	ts.log.Infof("schedule tx batch start, size %d", txBatchSize)

	var goRoutinePool *ants.Pool
	var err error
	poolCapacity := ts.StoreHelper.GetPoolCapacity()
	if goRoutinePool, err = ants.NewPool(poolCapacity, ants.WithPreAlloc(false)); err != nil {
		return nil, nil, err
	}
	defer goRoutinePool.Release()

	timeoutC := time.After(ScheduleTimeout * time.Second)
	startTime := time.Now()

	enableConflictsBitWindow, enableSenderGroup, conflictsBitWindow, senderGroup := ts.initOptimizeTools(txBatch)

	runningTxC := make(chan *commonPb.Transaction, txBatchSize)
	finishC := make(chan bool)
	if enableSenderGroup {
		if enableConflictsBitWindow {
			conflictsBitWindow.setMaxPoolCapacity(len(senderGroup.txsMap))
		}
		goRoutinePool.Tune(len(senderGroup.txsMap))
		go func() {
			ts.sendTxBySenderGroup(conflictsBitWindow, senderGroup, runningTxC, enableConflictsBitWindow)
		}()
	} else {
		go func() {
			if len(txBatch) > 0 {
				for _, tx := range txBatch {
					runningTxC <- tx
				}
			} else {
				finishC <- true
			}
		}()
	}
	// Put the pending transaction into the running queue
	go func() {
		for {
			select {
			case tx := <-runningTxC:
				ts.log.Debugf("prepare to submit running task for tx id:%s", tx.Payload.GetTxId())
				err := goRoutinePool.Submit(func() {
					// If snapshot is sealed, no more transaction will be added into snapshot
					if snapshot.IsSealed() {
						return
					}
					var start time.Time
					if localconf.ChainMakerConfig.MonitorConfig.Enabled {
						start = time.Now()
					}
					txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
					tx.Result = txSimContext.GetTxResult()

					// Apply failed means this tx's read set conflict with other txs' write set
					applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType,
						runVmSuccess, false)
					if !applyResult {
						if enableConflictsBitWindow {
							ts.adjustPoolSize(goRoutinePool, conflictsBitWindow, ConflictTx)
						}
						runningTxC <- tx
						ts.log.Debugf("apply to snapshot failed, tx id:%s, result:%+v, apply count:%d",
							tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize)
					} else {
						ts.handleApplyResult(enableConflictsBitWindow, enableSenderGroup,
							conflictsBitWindow, senderGroup, goRoutinePool, tx, start)
						ts.log.Debugf("apply to snapshot success, tx id:%s, result:%+v, apply count:%d",
							tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize)
					}
					// If all transactions have been successfully added to dag
					if applySize >= txBatchSize {
						finishC <- true
					}
				})
				if err != nil {
					ts.log.Warnf("failed to submit running task, tx id:%s during schedule, %+v",
						tx.Payload.GetTxId(), err)
				}
			case <-timeoutC:
				ts.scheduleFinishC <- true
				if enableSenderGroup {
					senderGroup.doneTxKeyC <- [32]byte{}
				}
				ts.log.Warnf("block [%d] schedule reached time limit", block.Header.BlockHeight)
				return
			case <-finishC:
				ts.log.Debugf("schedule finish")
				ts.scheduleFinishC <- true
				if enableSenderGroup {
					senderGroup.doneTxKeyC <- [32]byte{}
				}
				return
			}
		}
	}()

	// Wait for schedule finish signal
	<-ts.scheduleFinishC
	// Build DAG from read-write table
	snapshot.Seal()
	timeCostA := time.Since(startTime)
	block.Dag = snapshot.BuildDAG(ts.chainConf.ChainConfig().Contract.EnableSqlSupport)

	// Execute special tx sequentially, and add to dag
	if len(snapshot.GetSpecialTxTable()) > 0 {
		ts.simulateSpecialTxs(block.Dag, snapshot, block, txBatchSize)
	}

	timeCostB := time.Since(startTime)
	ts.log.Infof("schedule tx batch finished, success %d, txs execution cost %v, "+
		"dag building cost %v, total used %v, tps %v\n", len(block.Dag.Vertexes), timeCostA,
		timeCostB-timeCostA, timeCostB, float64(len(block.Dag.Vertexes))/(float64(timeCostB)/1e9))

	txRWSetMap := ts.getTxRWSetTable(snapshot, block)
	contractEventMap := ts.getContractEventMap(block)
	return txRWSetMap, contractEventMap, nil
}

func (ts *TxScheduler) initOptimizeTools(txBatch []*commonPb.Transaction) (bool, bool,
	*ConflictsBitWindow, *SenderGroup) {
	txBatchSize := len(txBatch)
	var conflictsBitWindow *ConflictsBitWindow
	var senderGroup *SenderGroup
	enableConflictsBitWindow := ts.chainConf.ChainConfig().Core.EnableConflictsBitWindow
	enableSenderGroup := ts.chainConf.ChainConfig().Core.EnableSenderGroup

	ts.log.Infof("enable conflicts bit window: [%t], enable sender group: [%t]\n",
		enableConflictsBitWindow, enableSenderGroup)

	if AdjustWindowSize*MinAdjustTimes > txBatchSize {
		enableConflictsBitWindow = false
	}
	if enableConflictsBitWindow {
		conflictsBitWindow = NewConflictsBitWindow(txBatchSize)
	}
	if enableSenderGroup {
		senderGroup = NewSenderGroup(txBatch)
	}
	return enableConflictsBitWindow, enableSenderGroup, conflictsBitWindow, senderGroup
}

func (ts *TxScheduler) sendTxBySenderGroup(conflictsBitWindow *ConflictsBitWindow, senderGroup *SenderGroup,
	runningTxC chan *commonPb.Transaction, enableConflictsBitWindow bool) {
	// first round
	for _, v := range senderGroup.txsMap {
		runningTxC <- v[0]
	}
	// solve done tx channel
	for {
		k := <-senderGroup.doneTxKeyC
		if k == [32]byte{} {
			return
		}
		senderGroup.txsMap[k] = senderGroup.txsMap[k][1:]
		if len(senderGroup.txsMap[k]) > 0 {
			runningTxC <- senderGroup.txsMap[k][0]
		} else {
			delete(senderGroup.txsMap, k)
			if enableConflictsBitWindow {
				conflictsBitWindow.setMaxPoolCapacity(len(senderGroup.txsMap))
			}
		}
	}
}

func (ts *TxScheduler) handleApplyResult(enableConflictsBitWindow bool, enableSenderGroup bool,
	conflictsBitWindow *ConflictsBitWindow, senderGroup *SenderGroup, goRoutinePool *ants.Pool,
	tx *commonPb.Transaction, start time.Time) {
	if enableConflictsBitWindow {
		ts.adjustPoolSize(goRoutinePool, conflictsBitWindow, NormalTx)
	}
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		elapsed := time.Since(start)
		ts.metricVMRunTime.WithLabelValues(tx.Payload.ChainId).Observe(elapsed.Seconds())
	}
	if enableSenderGroup {
		hashKey, _ := getSenderHashKey(tx)
		senderGroup.doneTxKeyC <- hashKey
	}
}

func (ts *TxScheduler) getTxRWSetTable(snapshot protocol.Snapshot, block *commonPb.Block) map[string]*commonPb.TxRWSet {
	txRWSetMap := make(map[string]*commonPb.TxRWSet)
	block.Txs = snapshot.GetTxTable()
	txRWSetTable := snapshot.GetTxRWSetTable()
	for _, txRWSet := range txRWSetTable {
		if txRWSet != nil {
			txRWSetMap[txRWSet.TxId] = txRWSet
		}
	}
	//ts.dumpDAG(block.Dag, block.Txs)
	if localconf.ChainMakerConfig.SchedulerConfig.RWSetLog {
		result, _ := prettyjson.Marshal(txRWSetMap)
		ts.log.Infof("schedule rwset :%s", result)
	}
	return txRWSetMap
}

func (ts *TxScheduler) getContractEventMap(block *commonPb.Block) map[string][]*commonPb.ContractEvent {
	contractEventMap := make(map[string][]*commonPb.ContractEvent)
	for _, tx := range block.Txs {
		event := tx.Result.ContractResult.ContractEvent
		contractEventMap[tx.Payload.TxId] = event
	}
	return contractEventMap
}

// SimulateWithDag based on the dag in the block, perform scheduling and execution transactions
func (ts *TxScheduler) SimulateWithDag(block *commonPb.Block, snapshot protocol.Snapshot) (
	map[string]*commonPb.TxRWSet, map[string]*commonPb.Result, error) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	var (
		startTime  = time.Now()
		txRWSetMap = make(map[string]*commonPb.TxRWSet)
	)
	if len(block.Txs) == 0 {
		ts.log.Debugf("no txs in block[%x] when simulate", block.Header.BlockHash)
		return txRWSetMap, snapshot.GetTxResultMap(), nil
	}
	ts.log.Infof("simulate with dag start, size %d", len(block.Txs))
	txMapping := make(map[int]*commonPb.Transaction)
	for index, tx := range block.Txs {
		txMapping[index] = tx
	}

	// Construct the adjacency list of dag, which describes the subsequent adjacency transactions of all transactions
	dag := block.Dag
	txIndexBatch, dagRemain, reverseDagRemain := ts.initSimulateDagGraph(dag)

	txBatchSize := len(block.Dag.Vertexes)
	if txBatchSize == 0 {
		ts.log.Error("found empty block when simulating txs")
		return nil, nil, fmt.Errorf("found empty block when simulating txs")
	}
	runningTxC := make(chan int, txBatchSize)
	doneTxC := make(chan int, txBatchSize)

	timeoutC := time.After(ScheduleWithDagTimeout * time.Second)
	finishC := make(chan bool)

	var goRoutinePool *ants.Pool
	var err error
	if goRoutinePool, err = ants.NewPool(len(block.Txs), ants.WithPreAlloc(true)); err != nil {
		return nil, nil, err
	}
	defer goRoutinePool.Release()

	ts.log.Debugf("block [%d] simulate with dag first batch size:%d, total batch size:%d",
		block.Header.BlockHeight, len(txIndexBatch), txBatchSize)

	go func() {
		for _, tx := range txIndexBatch {
			runningTxC <- tx
		}
	}()
	go func() {
		for {
			select {
			case txIndex := <-runningTxC:
				tx := txMapping[txIndex]
				ts.log.Debugf("simulate with dag, prepare to submit running task for tx id:%s", tx.Payload.GetTxId())
				err := goRoutinePool.Submit(func() {
					txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
					// if apply failed means this tx's read set conflict with other txs' write set
					applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType,
						runVmSuccess, true)
					if !applyResult {
						ts.log.Debugf("failed to apply snapshot for tx id:%s ", tx.Payload.TxId)
						runningTxC <- txIndex
					} else {
						ts.log.Debugf("apply to snapshot for tx id:%s, result:%+v, apply count:%d, tx batch size:%d",
							tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize, txBatchSize)
						doneTxC <- txIndex
					}
					// If all transactions in current batch have been successfully added to dag
					if applySize >= txBatchSize {
						ts.log.Debugf("finished 1 batch, apply size:%d, tx batch size:%d, dagRemain size:%d",
							applySize, txBatchSize, len(dagRemain))
						finishC <- true
					}
				})
				if err != nil {
					ts.log.Warnf("failed to submit tx id %s during simulate with dag, %+v",
						tx.Payload.GetTxId(), err)
				}
			case doneTxIndex := <-doneTxC:
				txIndexBatchAfterShrink := ts.shrinkDag(doneTxIndex, dagRemain, reverseDagRemain)
				ts.log.Debugf("block [%d] simulate with dag, pop next tx index batch size:%d, dagRemain size:%d",
					block.Header.BlockHeight, len(txIndexBatchAfterShrink), len(dagRemain))
				for _, tx := range txIndexBatchAfterShrink {
					runningTxC <- tx
				}
			case <-finishC:
				ts.log.Debugf("block [%d] simulate with dag finish", block.Header.BlockHeight)
				ts.scheduleFinishC <- true
				return
			case <-timeoutC:
				ts.log.Errorf("block [%d] simulate with dag timeout", block.Header.BlockHeight)
				ts.scheduleFinishC <- true
				return
			}
		}
	}()

	<-ts.scheduleFinishC
	snapshot.Seal()
	timeUsed := time.Since(startTime)
	ts.log.Infof("simulate with dag finished, size %d, time used %v, tps %v\n", len(block.Txs),
		timeUsed, float64(len(block.Txs))/(float64(timeUsed)/1e9))

	// Return the read and write set after the scheduled execution
	for _, txRWSet := range snapshot.GetTxRWSetTable() {
		if txRWSet != nil {
			txRWSetMap[txRWSet.TxId] = txRWSet
		}
	}
	if localconf.ChainMakerConfig.SchedulerConfig.RWSetLog {
		result, _ := prettyjson.Marshal(txRWSetMap)
		ts.log.Infof("simulate with dag rwset :%s", result)
	}
	return txRWSetMap, snapshot.GetTxResultMap(), nil
}

func (ts *TxScheduler) initSimulateDagGraph(dag *commonPb.DAG) ([]int, map[int]dagNeighbors, map[int]dagNeighbors) {
	dagRemain := make(map[int]dagNeighbors)
	reverseDagRemain := make(map[int]dagNeighbors)
	var txIndexBatch []int
	for txIndex, neighbors := range dag.Vertexes {
		if len(neighbors.Neighbors) == 0 {
			txIndexBatch = append(txIndexBatch, txIndex)
			continue
		}
		dn := make(dagNeighbors)
		for _, neighbor := range neighbors.Neighbors {
			dn[int(neighbor)] = true
			if _, ok := reverseDagRemain[int(neighbor)]; !ok {
				reverseDagRemain[int(neighbor)] = make(dagNeighbors)
			}
			reverseDagRemain[int(neighbor)][txIndex] = true
		}
		dagRemain[txIndex] = dn
	}
	return txIndexBatch, dagRemain, reverseDagRemain
}

func (ts *TxScheduler) adjustPoolSize(pool *ants.Pool, conflictsBitWindow *ConflictsBitWindow, txExecType TxExecType) {
	newPoolSize := conflictsBitWindow.Enqueue(txExecType, pool.Cap())
	if newPoolSize == -1 {
		return
	}
	pool.Tune(newPoolSize)
}

func (ts *TxScheduler) executeTx(tx *commonPb.Transaction, snapshot protocol.Snapshot, block *commonPb.Block) (
	protocol.TxSimContext, protocol.ExecOrderTxType, bool) {
	ts.log.Debugf("run vm start for tx:%s", tx.Payload.GetTxId())
	txSimContext := vm.NewTxSimContext(ts.VmManager, snapshot, tx, block.Header.BlockVersion, ts.log)
	ts.log.Debugf("new tx simulate context finished for tx id:%s", tx.Payload.GetTxId())
	runVmSuccess := true
	var txResult *commonPb.Result
	var err error
	var specialTxType protocol.ExecOrderTxType
	if txResult, specialTxType, err = ts.runVM(tx, txSimContext); err != nil {
		runVmSuccess = false
		ts.log.Errorf("failed to run vm for tx id:%s, tx result:%+v, error:%+v",
			tx.Payload.GetTxId(), txResult, err)
	}
	ts.log.Debugf("run vm finished for tx:%s, runVmSuccess:%v", tx.Payload.TxId, runVmSuccess)
	txSimContext.SetTxResult(txResult)
	return txSimContext, specialTxType, runVmSuccess
}

func (ts *TxScheduler) simulateSpecialTxs(dag *commonPb.DAG, snapshot protocol.Snapshot, block *commonPb.Block,
	txBatchSize int) {
	specialTxs := snapshot.GetSpecialTxTable()
	specialTxsLen := len(specialTxs)
	var firstTx *commonPb.Transaction
	runningTxC := make(chan *commonPb.Transaction, specialTxsLen)
	scheduleFinishC := make(chan bool)
	timeoutC := time.After(ScheduleWithDagTimeout * time.Second)
	go func() {
		for _, tx := range specialTxs {
			runningTxC <- tx
		}
	}()

	go func() {
		for {
			select {
			case tx := <-runningTxC:
				// simulate tx
				txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
				tx.Result = txSimContext.GetTxResult()
				// apply tx
				applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType, runVmSuccess, true)
				if !applyResult {
					ts.log.Debugf("failed to apply according to dag with tx %s ", tx.Payload.TxId)
					runningTxC <- tx
					continue
				}
				if firstTx == nil {
					firstTx = tx
					dagNeighbors := &commonPb.DAG_Neighbor{
						Neighbors: make([]uint32, 0, snapshot.GetSnapshotSize()-1),
					}
					for i := uint32(0); i < uint32(snapshot.GetSnapshotSize()-1); i++ {
						dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, i)
					}
					dag.Vertexes = append(dag.Vertexes, dagNeighbors)
				} else {
					dagNeighbors := &commonPb.DAG_Neighbor{
						Neighbors: make([]uint32, 0, 1),
					}
					dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, uint32(snapshot.GetSnapshotSize())-2)
					dag.Vertexes = append(dag.Vertexes, dagNeighbors)
				}
				if applySize >= txBatchSize {
					ts.log.Debugf("block [%d] schedule special txs finished, apply size:%d, len of txs:%d, "+
						"len of special txs:%d", block.Header.BlockHeight, applySize, txBatchSize, specialTxsLen)
					scheduleFinishC <- true
					return
				}
			case <-timeoutC:
				ts.log.Errorf("block [%d] schedule special txs timeout", block.Header.BlockHeight)
				scheduleFinishC <- true
				return
			}
		}
	}()
	<-scheduleFinishC
}

func (ts *TxScheduler) shrinkDag(txIndex int, dagRemain map[int]dagNeighbors,
	reverseDagRemain map[int]dagNeighbors) []int {
	var txIndexBatch []int
	for k := range reverseDagRemain[txIndex] {
		delete(dagRemain[k], txIndex)
		if len(dagRemain[k]) == 0 {
			txIndexBatch = append(txIndexBatch, k)
			delete(dagRemain, k)
		}
	}
	delete(reverseDagRemain, txIndex)
	return txIndexBatch
}

func (ts *TxScheduler) Halt() {
	ts.scheduleFinishC <- true
}

func (ts *TxScheduler) runVM(tx *commonPb.Transaction, txSimContext protocol.TxSimContext) (
	*commonPb.Result, protocol.ExecOrderTxType, error) {
	var (
		contractName          string
		method                string
		byteCode              []byte
		pk                    []byte
		specialTxType         protocol.ExecOrderTxType
		accountMangerContract *commonPb.Contract
		contractResultPayload *commonPb.ContractResult
		txStatusCode          commonPb.TxStatusCode
	)

	result := &commonPb.Result{
		Code: commonPb.TxStatusCode_SUCCESS,
		ContractResult: &commonPb.ContractResult{
			Code:    uint32(0),
			Result:  nil,
			Message: "",
		},
		RwSetHash: nil,
	}
	payload := tx.Payload
	if payload.TxType != commonPb.TxType_QUERY_CONTRACT && payload.TxType != commonPb.TxType_INVOKE_CONTRACT {
		return errResult(result, fmt.Errorf("no such tx type: %s", tx.Payload.TxType))
	}

	contractName = payload.ContractName
	method = payload.Method
	parameters, err := ts.parseParameter(payload.Parameters)
	if err != nil {
		ts.log.Errorf("parse contract[%s] parameters error:%s", contractName, err)
		return errResult(result, fmt.Errorf(
			"parse tx[%s] contract[%s] parameters error:%s",
			payload.TxId,
			contractName,
			err.Error()),
		)
	}

	contract, err := txSimContext.GetContractByName(contractName)
	if err != nil {
		ts.log.Errorf("Get contract info by name[%s] error:%s", contractName, err)
		return errResult(result, err)
	}
	if contract.RuntimeType != commonPb.RuntimeType_NATIVE && contract.RuntimeType != commonPb.RuntimeType_DOCKER_GO {
		byteCode, err = txSimContext.GetContractBytecode(contractName)
		if err != nil {
			ts.log.Errorf("Get contract bytecode by name[%s] error:%s", contractName, err)
			return errResult(result, err)
		}
	} else {
		ts.log.DebugDynamic(func() string {
			contractData, _ := json.Marshal(contract)
			return fmt.Sprintf("contract[%s] is a native contract, definition:%s",
				contractName, string(contractData))
		})
	}

	accountMangerContract, pk, err = ts.getAccountMgrContractAndPk(txSimContext, tx, contractName, method)
	if err != nil {
		return result, specialTxType, err
	}

	// charge gas limit
	_, err = ts.chargeGasLimit(accountMangerContract, tx, txSimContext, contractName, method, pk, result)
	if err != nil {
		ts.log.Errorf("charge gas limit err is %v", err)
		result.Code = commonPb.TxStatusCode_GAS_BALANCE_NOT_ENOUGH_FAILED
		result.Message = err.Error()
		result.ContractResult.Code = uint32(1)
		result.ContractResult.Message = err.Error()
		return result, specialTxType, err
	}

	contractResultPayload, specialTxType, txStatusCode = ts.VmManager.RunContract(contract, method, byteCode,
		parameters, txSimContext, 0, tx.Payload.TxType)
	result.Code = txStatusCode
	result.ContractResult = contractResultPayload

	// refund gas
	_, err = ts.refundGas(accountMangerContract, tx, txSimContext, contractName, method, pk, result,
		contractResultPayload)
	if err != nil {
		ts.log.Errorf("refund gas err is %v", err)
	}

	if txStatusCode == commonPb.TxStatusCode_SUCCESS {
		return result, specialTxType, nil
	}
	return result, specialTxType, errors.New(contractResultPayload.Message)
}
func errResult(result *commonPb.Result, err error) (*commonPb.Result, protocol.ExecOrderTxType, error) {
	result.ContractResult.Message = err.Error()
	result.Code = commonPb.TxStatusCode_INVALID_PARAMETER
	result.ContractResult.Code = 1
	return result, protocol.ExecOrderTxTypeNormal, err
}
func (ts *TxScheduler) parseParameter(parameterPairs []*commonPb.KeyValuePair) (map[string][]byte, error) {
	// verify parameters
	if len(parameterPairs) > protocol.ParametersKeyMaxCount {
		return nil, fmt.Errorf(
			"expect parameters length less than %d, but got %d",
			protocol.ParametersKeyMaxCount,
			len(parameterPairs),
		)
	}
	parameters := make(map[string][]byte, 16)
	for i := 0; i < len(parameterPairs); i++ {
		key := parameterPairs[i].Key
		value := parameterPairs[i].Value
		if len(key) > protocol.DefaultMaxStateKeyLen {
			return nil, fmt.Errorf(
				"expect key length less than %d, but got %d",
				protocol.DefaultMaxStateKeyLen,
				len(key),
			)
		}
		match := ts.keyReg.MatchString(key)
		if !match {
			return nil, fmt.Errorf(
				"expect key no special characters, but got key:[%s]. letter, number, dot and underline are allowed",
				key,
			)
		}
		if len(value) > int(protocol.ParametersValueMaxLength) {
			return nil, fmt.Errorf(
				"expect value length less than %d, but got %d",
				protocol.ParametersValueMaxLength,
				len(value),
			)
		}

		parameters[key] = value
	}
	return parameters, nil
}

//nolint: unused
func (ts *TxScheduler) dumpDAG(dag *commonPb.DAG, txs []*commonPb.Transaction) {
	dagString := "digraph DAG {\n"
	for i, ns := range dag.Vertexes {
		if len(ns.Neighbors) == 0 {
			dagString += fmt.Sprintf("id_%s -> begin;\n", txs[i].Payload.TxId[:8])
			continue
		}
		for _, n := range ns.Neighbors {
			dagString += fmt.Sprintf("id_%s -> id_%s;\n", txs[i].Payload.TxId[:8], txs[n].Payload.TxId[:8])
		}
	}
	dagString += "}"
	ts.log.Infof("Dump Dag: %s", dagString)
}

func (ts *TxScheduler) chargeGasLimit(accountMangerContract *commonPb.Contract, tx *commonPb.Transaction,
	txSimContext protocol.TxSimContext, contractName, method string, pk []byte,
	result *commonPb.Result) (re *commonPb.Result, err error) {
	if ts.checkGasEnable() && ts.checkNativeFilter(contractName, method) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		var code commonPb.TxStatusCode
		var runChargeGasContract *commonPb.ContractResult
		var limit uint64
		if tx.Payload.Limit == nil {
			err = errors.New("tx payload limit is nil")
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		limit = tx.Payload.Limit.GasLimit
		chargeParameters := map[string][]byte{
			accountmgr.ChargePublicKey: pk,
			accountmgr.ChargeGasAmount: []byte(strconv.FormatUint(limit, 10)),
		}
		runChargeGasContract, _, code = ts.VmManager.RunContract(
			accountMangerContract, syscontract.GasAccountFunction_CHARGE_GAS.String(),
			nil, chargeParameters, txSimContext, 0, commonPb.TxType_INVOKE_CONTRACT)
		if code != commonPb.TxStatusCode_SUCCESS {
			result.Code = code
			result.ContractResult = runChargeGasContract
			return result, errors.New(runChargeGasContract.Message)
		}
	}
	return result, nil
}

func (ts *TxScheduler) refundGas(accountMangerContract *commonPb.Contract, tx *commonPb.Transaction,
	txSimContext protocol.TxSimContext, contractName, method string, pk []byte,
	result *commonPb.Result, contractResultPayload *commonPb.ContractResult) (re *commonPb.Result, err error) {
	if ts.checkGasEnable() && ts.checkNativeFilter(contractName, method) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		var code commonPb.TxStatusCode
		var refundGasContract *commonPb.ContractResult
		var limit uint64
		if tx.Payload.Limit == nil {
			err = errors.New("tx payload limit is nil")
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		limit = tx.Payload.Limit.GasLimit
		if limit < contractResultPayload.GasUsed {
			err = fmt.Errorf("gas limit is not enough, [limit:%d]/[gasUsed:%d]", limit, contractResultPayload.GasUsed)
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		refundGas := limit - contractResultPayload.GasUsed
		ts.log.Debugf("refund gas [%d], gas used [%d]", refundGas, contractResultPayload.GasUsed)

		if refundGas == 0 {
			return result, nil
		}

		refundGasParameters := map[string][]byte{
			accountmgr.RechargeKey:       pk,
			accountmgr.RechargeAmountKey: []byte(strconv.FormatUint(refundGas, 10)),
		}

		refundGasContract, _, code = ts.VmManager.RunContract(
			accountMangerContract, syscontract.GasAccountFunction_REFUND_GAS_VM.String(),
			nil, refundGasParameters, txSimContext, 0, commonPb.TxType_INVOKE_CONTRACT)
		if code != commonPb.TxStatusCode_SUCCESS {
			result.Code = code
			result.ContractResult = refundGasContract
			return result, errors.New(refundGasContract.Message)
		}
	}
	return result, nil
}

func (ts *TxScheduler) getAccountMgrContractAndPk(txSimContext protocol.TxSimContext, tx *commonPb.Transaction,
	contractName, method string) (accountMangerContract *commonPb.Contract, pk []byte, err error) {
	if ts.checkGasEnable() && ts.checkNativeFilter(contractName, method) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		accountMangerContract, err = txSimContext.GetContractByName(syscontract.SystemContract_ACCOUNT_MANAGER.String())
		if err != nil {
			ts.log.Error(err.Error())
			return nil, nil, err
		}

		pk, err = ts.getSenderPk(txSimContext)
		if err != nil {
			ts.log.Error(err.Error())
			return accountMangerContract, nil, err
		}
		return accountMangerContract, pk, err
	}
	return nil, nil, nil
}

func (ts *TxScheduler) checkGasEnable() bool {
	if ts.chainConf.ChainConfig() != nil && ts.chainConf.ChainConfig().AccountConfig != nil {
		ts.log.Debugf("chain config account config enable gas is:%v", ts.chainConf.ChainConfig().AccountConfig.EnableGas)
		return ts.chainConf.ChainConfig().AccountConfig.EnableGas
	}
	return false
}

func (ts *TxScheduler) checkNativeFilter(contractName, method string) bool {
	if !utils.IsNativeContract(contractName) {
		return true
	}
	if method == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
		method == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
		return true
	}
	return false
}

func (ts *TxScheduler) getSenderPk(txSimContext protocol.TxSimContext) ([]byte, error) {

	var err error
	var pk []byte
	sender := txSimContext.GetSender()
	if sender == nil {
		err = errors.New(" can not find sender from tx ")
		ts.log.Error(err.Error())
		return nil, err
	}

	switch sender.MemberType {
	case accesscontrol.MemberType_CERT:
		pk, err = publicKeyFromCert(sender.MemberInfo)
		if err != nil {
			ts.log.Error(err.Error())
			return nil, err
		}
	case accesscontrol.MemberType_CERT_HASH:
		var certInfo *commonPb.CertInfo
		infoHex := hex.EncodeToString(sender.MemberInfo)
		if certInfo, err = wholeCertInfo(txSimContext, infoHex); err != nil {
			ts.log.Error(err.Error())
			return nil, fmt.Errorf(" can not load the whole cert info,member[%s],reason: %s", infoHex, err)
		}

		if pk, err = publicKeyFromCert(certInfo.Cert); err != nil {
			ts.log.Error(err.Error())
			return nil, err
		}

	case accesscontrol.MemberType_PUBLIC_KEY:
		pk = sender.MemberInfo
	default:
		err = fmt.Errorf("invalid member type: %s", sender.MemberType)
		ts.log.Error(err.Error())
		return nil, err
	}

	return pk, nil
}

// parseUserAddress
func publicKeyFromCert(member []byte) ([]byte, error) {
	certificate, err := utils.ParseCert(member)
	if err != nil {
		return nil, err
	}
	pubKeyBytes, err := certificate.PublicKey.Bytes()
	if err != nil {
		return nil, err
	}
	return pubKeyBytes, nil
}

func wholeCertInfo(txSimContext protocol.TxSimContext, certHash string) (*commonPb.CertInfo, error) {
	certBytes, err := txSimContext.Get(syscontract.SystemContract_CERT_MANAGE.String(), []byte(certHash))
	if err != nil {
		return nil, err
	}

	return &commonPb.CertInfo{
		Hash: certHash,
		Cert: certBytes,
	}, nil
}

type SenderGroup struct {
	txsMap     map[[32]byte][]*commonPb.Transaction
	doneTxKeyC chan [32]byte
}

func NewSenderGroup(txBatch []*commonPb.Transaction) *SenderGroup {
	return &SenderGroup{
		txsMap:     getSenderTxsMap(txBatch),
		doneTxKeyC: make(chan [32]byte, len(txBatch)),
	}
}

func getSenderTxsMap(txBatch []*commonPb.Transaction) map[[32]byte][]*commonPb.Transaction {
	senderTxsMap := make(map[[32]byte][]*commonPb.Transaction)
	for _, tx := range txBatch {
		hashKey, _ := getSenderHashKey(tx)
		senderTxsMap[hashKey] = append(senderTxsMap[hashKey], tx)
	}
	return senderTxsMap
}

func getSenderHashKey(tx *commonPb.Transaction) ([32]byte, error) {
	sender := tx.GetSender().GetSigner()
	keyBytes, err := sender.Marshal()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(keyBytes), nil
}
