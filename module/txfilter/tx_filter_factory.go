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

func (cf *txFilterFactory) NewTxFilter(conf *config.TxFilterConfig, log protocol.Logger,
	store protocol.BlockchainStore) (protocol.TxFilter, error) {
	if conf == nil {
		log.Warn("txfilter conf is nil, use default: store")
		return defau1t.Init(store), nil
	}
	switch conf.Type {
	case config.TxFilterType_None:
		return defau1t.Init(store), nil
	case config.TxFilterType_BirdsNest:
		return birdnest.Init(conf.BirdsNest, log, store)
	case config.TxFilterType_Map:
		return mapimpl.Init(), nil
	case config.TxFilterType_ShardingBirdsNest:
		return shardingbirdsnest.Init(conf.ShardingBirdsNest, log, store)
	default:
		log.Warnf("txfilter type(%v) not support, use default: store", conf.Type)
		return defau1t.Init(store), nil
	}
}
