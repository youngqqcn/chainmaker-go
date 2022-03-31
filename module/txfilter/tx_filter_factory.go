/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package txfilter

import (
	"sync"

	mapimpl "chainmaker.org/chainmaker-go/module/txfilter/map"

	"chainmaker.org/chainmaker-go/module/txfilter/birdnest"
	"chainmaker.org/chainmaker-go/module/txfilter/defau1t"
	"chainmaker.org/chainmaker-go/module/txfilter/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
)

// txFilterFactory Transaction filter factory
type txFilterFactory struct {
}

var once sync.Once
var _instance *txFilterFactory

// Factory return the global tx filter factory.
//nolint: revive
func Factory() *txFilterFactory {
	once.Do(func() { _instance = new(txFilterFactory) })
	return _instance
}

// NewTxFilter new transaction filter
func (cf *txFilterFactory) NewTxFilter(conf *config.TxFilterConfig, log protocol.Logger,
	store protocol.BlockchainStore) (protocol.TxFilter, error) {
	if conf == nil {
		log.Warn("txfilter conf is nil, use default type: store")
		return defau1t.New(store), nil
	}
	switch conf.Type {
	case config.TxFilterType_None:
		return defau1t.New(store), nil
	case config.TxFilterType_BirdsNest:
		return birdnest.New(conf.BirdsNest, log, store)
	case config.TxFilterType_Map:
		return mapimpl.New(), nil
	case config.TxFilterType_ShardingBirdsNest:
		return shardingbirdsnest.New(conf.ShardingBirdsNest, log, store)
	default:
		log.Warnf("txfilter type: %v not support, use default type: store", conf.Type)
		return defau1t.New(store), nil
	}
}
