/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package shardingbirdsnest

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
)

const (
	shardingLogTemplate = "sharding[%d]%v "
)

type TxFilter struct {
	log   protocol.Logger
	bn    *sbn.ShardingBirdsNest
	store protocol.BlockchainStore
	exitC chan struct{}
}

func Init(config *common.ShardingBirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (
	protocol.TxFilter, error) {
	init := utils.CurrentTimeMillisSeconds()
	exitC := make(chan struct{})
	shardingBirdsNest, err := sbn.NewShardingBirdsNest(config, exitC, bn.LruPolicy, sbn.NewModuloSA(int(config.Length)),
		filtercommon.NewLogger(log))
	if err != nil {
		log.Errorf("new sharding bird's nest error: %v", err)
		return nil, err
	}
	txFilter := &TxFilter{
		log:   log,
		bn:    shardingBirdsNest,
		exitC: exitC,
		store: store,
	}
	init = utils.CurrentTimeMillisSeconds() - init

	chase := utils.CurrentTimeMillisSeconds()
	// Chase block height
	err = filtercommon.ChaseBlockHeight(store, txFilter, log)
	if err != nil {
		log.Errorf("chase block height error: %v", err)
		return nil, err
	}
	chase = utils.CurrentTimeMillisSeconds() - chase
	log.DebugDynamic(func() string {
		return filtercommon.LoggingFixLength("init success sharding %v, bird's nest %v max keys %v, time: %v, "+
			"chase: %v", config.Length, config.Birdsnest.Length, config.Birdsnest.Cuckoo.MaxNumKeys, init+chase, chase)
	})
	shardingBirdsNest.Start()
	return txFilter, nil
}

func (f *TxFilter) GetHeight() uint64 {
	return f.bn.GetHeight()
}

func (f *TxFilter) SetHeight(height uint64) {
	f.bn.SetHeight(height)
}

func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...common.RuleType) (exists bool, height uint64,
	err error) {
	start := time.Now()
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		f.log.Warnf("TxId is not a timestamp key query store, txid: %v, err: %v", txId, err)
		exists, height, err = f.store.TxExistsInFullDB(txId)
		if err != nil {
			return false, height, err
		}
		return
	}
	contains, err := f.bn.Contains(key, ruleType...)
	if err != nil {
		f.log.Errorf("contains tx id:%v error: %v", txId, err)
		return false, f.GetHeight(), err
	}

	if contains {
		exists, err = f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("store tx id: %v error: %v", txId, err)
			return false, f.GetHeight(), err
		}
		elapsed := time.Since(start)
		// true or false positive
		f.log.DebugDynamic(func() string {
			return filtercommon.LoggingFixLength("positive: %v, tx id: %v, time: %v", exists, txId, elapsed)
		})
		return exists, f.GetHeight(), nil
	}
	elapsed := time.Since(start)

	f.log.DebugDynamic(func() string {
		return filtercommon.LoggingFixLength("contains %v tx id:%v time: %v,", contains, txId, elapsed)
	})
	return contains, f.GetHeight(), nil
}

func (f *TxFilter) Add(txId string) error {
	add := utils.CurrentTimeMillisSeconds()
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		f.log.Warnf("this is normal key: %v", txId, err)
		return nil
	}
	err = f.bn.Add(key)
	if err != nil {
		f.log.Errorf("add error key: %v, error: %v", txId, err)
		return err
	}
	add -= utils.CurrentTimeMillisSeconds()
	f.log.DebugDynamic(func() string {
		return filtercommon.LoggingFixLength("add key: %v, time: %v", txId, -add)
	})
	return nil
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	start := time.Now()
	timestampKeys, _ := bn.S2timestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		return nil
	}
	err := f.bn.Adds(timestampKeys)
	if err != nil {
		f.log.Warnf("adds error keys: %v, error: %v", len(txIds), err)
		return err
	}
	elapsed := time.Since(start)
	f.log.DebugDynamic(func() string {
		return filtercommon.LoggingFixLength("adds success keys: %v, height: %v, time: %v infos:%v ", len(txIds),
			f.GetHeight(),
			elapsed.String(), func() string {
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
			}())
	})
	return nil
}

func (f *TxFilter) IsExists(txId string, ruleType ...common.RuleType) (exists bool, err error) {
	start := time.Now()
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		f.log.Warnf("TxId is not a timestamp key query store, txid: %v, err: %v", txId, err)
		exists, err = f.store.TxExists(txId)
		if err != nil {
			return false, err
		}
		return
	}
	contains, err := f.bn.Contains(key, ruleType...)
	if err != nil {
		f.log.Errorf("contains tx id:%v error: %v", txId, err)
		return false, err
	}

	if contains {
		exists, err = f.store.TxExists(txId)
		if err != nil {
			f.log.Errorf("store tx id: %v error: %v", txId, err)
			return false, err
		}
		elapsed := time.Since(start)
		// true or false positive
		f.log.DebugDynamic(func() string {
			return filtercommon.LoggingFixLength("positive: %v, tx id: %v, time: %v", exists, txId, elapsed)
		})
		return exists, nil
	}
	elapsed := time.Since(start)

	f.log.DebugDynamic(func() string {
		return filtercommon.LoggingFixLength("contains %v tx id:%v time: %v,", contains, txId, elapsed)
	})
	return contains, nil
}

func (f *TxFilter) Close() {
	close(f.exitC)
}
