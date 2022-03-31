/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package birdnest

import (
	"time"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// TxFilter bn.BirdsNestImpl transaction filter
type TxFilter struct {
	log   protocol.Logger
	bn    *bn.BirdsNestImpl
	store protocol.BlockchainStore
	exitC chan struct{}
}

// New transaction filter init
func New(config *commonPb.BirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (
	protocol.TxFilter, error) {
	initLasts := time.Now()
	exitC := make(chan struct{})
	birdnest, err := bn.NewBirdsNest(config, exitC, bn.LruStrategy, filtercommon.NewLogger(log))
	if err != nil {
		log.Errorf("new filter fail, error: %v", err)
		return nil, err
	}
	txFilter := &TxFilter{
		log:   log,
		bn:    birdnest,
		exitC: exitC,
	}
	err = filtercommon.ChaseBlockHeight(store, txFilter, log)
	if err != nil {
		return nil, err
	}
	log.Infof("bird's nest filter init success, size: %v, max keys: %v, cost: %v",
		config.Length, config.Cuckoo.MaxNumKeys, time.Since(initLasts))
	birdnest.Start()
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
func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...commonPb.RuleType) (bool, uint64, error) {
	exists, err := f.IsExists(txId, ruleType...)
	if err != nil {
		return false, 0, err
	}
	return exists, f.GetHeight(), nil
}

// Add txId to transaction filter
func (f *TxFilter) Add(txId string) error {
	timestampKey, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	return f.bn.Add(timestampKey)
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	start := time.Now()
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) > 0 {
		err := f.bn.Adds(timestampKeys)
		if err != nil {
			f.log.Warnf("filter adds fail, txid size: %v, error: %v", len(txIds), err)
		}
	}
	err := f.bn.Adds(timestampKeys)
	if err != nil {
		f.log.Errorf("filter adds fail, txid size: %v, error: %v", len(txIds), err)
		return err
	}

	f.addsPrintInfo(txIds, start)
	return nil
}

// index 1 cuckoo size
// index 2 current index
// index 3 total cuckoo size
// index 4 total space occupied by cuckoo

func (f *TxFilter) addsPrintInfo(txIds []string, start time.Time) {
	info := f.bn.Info()
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc(
		"filter adds success, height: %v, txids: %v, size: %v, curr: %v, total keys: %v, bytes: %v, cost: %v",
		f.GetHeight(), len(txIds), info[1], info[2], info[3], info[4],
		time.Since(start),
	))
}

// AddsAndSetHeight batch add tx id and set height
func (f *TxFilter) AddsAndSetHeight(txIds []string, height uint64) error {
	start := time.Now()
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		return nil
	}
	err := f.bn.AddsAndSetHeight(timestampKeys, height)
	if err != nil {
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f *TxFilter) IsExists(txId string, ruleType ...commonPb.RuleType) (exists bool, err error) {
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		exists, err = f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("filter check exists, query from db fail, normal txid: %v, error:%v", txId, err)
			return false, err
		}
		return exists, nil
	}
	contains, err := f.bn.Contains(key, ruleType...)
	if err != nil {
		f.log.Errorf("filter check exists, query from filter fail, txid: %v, error: %v", txId, err)
		return false, err
	}
	if contains {
		// False positive treatment
		exists, err = f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("filter check exists, query from db fail, txid: %v, error:%v", txId, err)
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return contains, nil
}

// Close transaction filter
func (f *TxFilter) Close() {
	close(f.exitC)
}
