/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package shardingbirdsnest

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

const (
	shardingLogTemplate = "sharding[%d]%v "
)

// TxFilter Sharding transaction filter
type TxFilter struct {
	log   protocol.Logger
	bn    *sbn.ShardingBirdsNest
	store protocol.BlockchainStore
	exitC chan struct{}
	l     sync.RWMutex
}

func (f *TxFilter) ValidateRule(txId string, ruleType ...common.RuleType) error {
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	err = f.bn.ValidateRule(key, ruleType...)
	if err != nil {
		return err
	}
	return nil
}

// New transaction filter init
func New(config *common.ShardingBirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (
	protocol.TxFilter, error) {
	// Because it is compatible with Normal type, the transaction ID cannot be converted to time transaction ID, so the
	// database can be queried directly. Therefore, the transaction ID type is fixed as TimestampKey
	config.Birdsnest.Cuckoo.KeyType = common.KeyType_KTTimestampKey

	initLasts := time.Now()
	exitC := make(chan struct{})
	shardingBirdsNest, err := sbn.NewShardingBirdsNest(config, exitC, bn.LruStrategy, sbn.NewModuloSA(int(config.Length)),
		filtercommon.NewLogger(log))
	if err != nil {
		log.Errorf("new filter fail, error: %v", err)
		if err != bn.ErrCannotModifyTheNestConfiguration {
			return nil, err
		}
	}
	txFilter := &TxFilter{
		log:   log,
		bn:    shardingBirdsNest,
		exitC: exitC,
		store: store,
	}
	shardingBirdsNest.Start()

	// Chase block height
	err = filtercommon.ChaseBlockHeight(store, txFilter, log)
	if err != nil {
		return nil, err
	}
	log.Infof("shading filter init success, sharding: %v, birdsnest: %v max keys: %v, cost: %v", config.Length,
		config.Birdsnest.Length, config.Birdsnest.Cuckoo.MaxNumKeys, time.Since(initLasts))
	return txFilter, nil
}

// GetHeight get height from transaction filter
func (f *TxFilter) GetHeight() uint64 {
	return f.bn.GetHeight()
}

// SetHeight set height from transaction filter
func (f *TxFilter) SetHeight(height uint64) {
	f.bn.SetHeight(height)
}

// IsExistsAndReturnHeight is exists and return height
func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...common.RuleType) (exists bool, height uint64,
	err error) {
	isExists, err := f.IsExists(txId, ruleType...)
	if err != nil {
		return false, 0, err
	}
	return isExists, f.GetHeight(), nil
}

// Add txId to transaction filter
func (f *TxFilter) Add(txId string) error {
	start := time.Now()
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	f.l.Lock()
	err = f.bn.Add(key)
	f.l.Unlock()
	if err != nil {
		f.log.Errorf("filter add fail, txid: %v error: %v", txId, err)
		return err
	}
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("filter add txid: %v, cost: %v", txId, time.Since(start)))
	return nil
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	start := time.Now()
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		return nil
	}
	f.l.Lock()
	err := f.bn.Adds(timestampKeys)
	f.l.Unlock()
	if err != nil {
		f.log.Errorf("filter adds fail, ids: %v, error: %v", len(txIds), err)
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

func (f *TxFilter) addsPrintInfo(txIds []string, start time.Time) {
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc(
		"filter adds success, ids: %v height: %v, cost: %v infos:%v ",
		len(txIds),
		f.GetHeight(),
		time.Since(start),
		func() string {
			var (
				bt    bytes.Buffer
				total uint64
			)
			for sharding, infos := range f.bn.Infos() {
				bt.WriteString(fmt.Sprintf(shardingLogTemplate, sharding, infos[3]))
				total += infos[3]
			}
			bt.WriteString("total:")
			bt.WriteString(strconv.FormatUint(total, 10))
			return bt.String()
		}(),
	))
}

// AddsAndSetHeight batch add tx id and set height
func (f *TxFilter) AddsAndSetHeight(txIds []string, height uint64) error {
	start := time.Now()
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		f.SetHeight(height)
		f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("adds and set height, no timestamp keys height: %d",
			height))
		return nil
	}
	f.l.Lock()
	err := f.bn.AddsAndSetHeight(timestampKeys, height)
	f.l.Unlock()
	if err != nil {
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f *TxFilter) IsExists(txId string, ruleType ...common.RuleType) (bool, error) {
	start := time.Now()
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		var exists bool
		exists, err = f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("filter check exists, query from db fail, normal txid: %v, error:%v", txId, err)
			return false, err
		}
		return exists, err
	}
	f.l.RLock()
	defer f.l.RUnlock()
	contains, err := f.bn.Contains(key, ruleType...)
	if err != nil {
		f.log.Errorf("filter check exists, query from filter fail, txid: %v, error:%v", txId, err)
		return false, err
	}

	if contains {
		exists, err := f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("filter check exists, query from db fail, txid: %v, error: %v", txId, err)
			return false, err
		}
		// true or false positive
		f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("filter check exists, %v positive txid: %v, "+
			"cost: %v", exists, txId, time.Since(start)))
		return exists, nil
	}

	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("filter check exists, false txid: %v, cost: %v,",
		txId, time.Since(start)))
	return contains, nil
}

// Close transaction filter
func (f *TxFilter) Close() {
	close(f.exitC)
}
