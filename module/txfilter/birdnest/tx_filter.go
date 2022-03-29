/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package birdnest

import (
	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

type TxFilter struct {
	log   protocol.Logger
	bn    *bn.BirdsNestImpl
	store protocol.BlockchainStore
	exitC chan struct{}
}

func Init(conf *commonPb.BirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (
	protocol.TxFilter, error) {
	exitC := make(chan struct{})
	birdnest, err := bn.NewBirdsNest(conf, exitC, bn.LruPolicy, filtercommon.NewLogger(log))
	if err != nil {
		log.Errorf("new bird's nest error: %v", err)
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
	birdnest.Start()
	return txFilter, nil
}

func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...commonPb.RuleType) (bool, uint64, error) {
	exists, err := f.IsExists(txId, ruleType...)
	if err != nil {
		return false, 0, err
	}
	return exists, f.GetHeight(), nil
}

func (f *TxFilter) GetHeight() uint64 {
	return f.bn.GetHeight()
}

func (f *TxFilter) SetHeight(height uint64) {
	f.bn.SetHeight(height)
}

func (f *TxFilter) Add(txId string) error {
	timestampKey, err := bn.ToTimestampKey(txId)
	if err != nil {
		f.log.Warnf("this is normal key: %v", txId, err)
		return nil
	}
	return f.bn.Add(timestampKey)
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	timestampKeys, _ := bn.S2timestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) > 0 {
		err := f.bn.Adds(timestampKeys)
		if err != nil {
			f.log.Warnf("adds error keys: %v, error: %v", len(txIds), err)
		}
	}
	err := f.bn.Adds(timestampKeys)
	if err != nil {
		f.log.Errorf("adds error keys: %v, error: %v", len(txIds), err)
		return err
	}
	infos := f.bn.Info()
	f.log.InfoDynamic(func() string {
		return filtercommon.LoggingFixLength("infos success height: %v, size: %v, curr: %v, keys: %v, Bytes: %v",
			f.GetHeight(), infos[1], infos[2], infos[3], infos[4])
	})
	return nil
}

func (f *TxFilter) IsExists(txId string, ruleType ...commonPb.RuleType) (exists bool, err error) {
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		f.log.Warnf("TxId is not a timestamp key query store, txid: %v, err: %v", txId, err)
		exists, err = f.store.TxExists(txId)
		if err != nil {
			return false, err
		}
		return exists, nil
	}
	contains, err := f.bn.Contains(key, ruleType...)
	if err != nil {
		f.log.Error(err)
	}
	if contains {
		// False positive treatment
		exists, err = f.store.TxExists(txId)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return contains, nil

}

func (f *TxFilter) Close() {
	close(f.exitC)
}
