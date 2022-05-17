/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package filtercommon

import (
	"errors"
	"fmt"
	"time"

	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/gogo/protobuf/proto"
)

// ChaseBlockHeight Chase high block
func ChaseBlockHeight(store protocol.BlockchainStore, filter protocol.TxFilter, log protocol.Logger) error {
	cost := time.Now()
	lastBlock, err := store.GetLastBlock()
	if err != nil {
		log.Errorf("query last block from db fail, error: %v", err)
		return err
	}
	log.Infof("chase block start,filter height: %v, block height: %v", filter.GetHeight(),
		lastBlock.Header.BlockHeight)
	for height := filter.GetHeight() + 1; height <= lastBlock.Header.BlockHeight; height++ {
		var block *common.Block
		if height != lastBlock.Header.BlockHeight {
			block, err = store.GetBlock(height)
			if err != nil {
				log.Errorf("query block from db fail, height: %v, error: %v", height, err)
				return err
			}
		} else {
			block = lastBlock
		}
		ids := utils.GetTxIds(block.Txs)
		err = filter.AddsAndSetHeight(ids, block.Header.BlockHeight)
		if err != nil {
			log.Errorf("chase block add fail, height: %v, keys: %v, error: %v", block.Header.BlockHeight, len(ids), err)
			return err
		}
		log.Infof("chasing block, height: %d", block.Header.BlockHeight)
	}
	log.Infof("chase block finish, height: %d, block height: %d, cost: %d", filter.GetHeight(),
		lastBlock.Header.BlockHeight, time.Since(cost))

	return nil
}

func GetConf(chainId string) (*config.TxFilterConfig, error) {
	return ToPbConfig(localconf.ChainMakerConfig.TxFilter, chainId)
}

// ToPbConfig Convert localconf.TxFilterConfig to config.TxFilterConfig
func ToPbConfig(conf localconf.TxFilterConfig, chainId string) (*config.TxFilterConfig, error) {
	c := &config.TxFilterConfig{
		Type: config.TxFilterType(conf.Type),
	}
	switch config.TxFilterType(conf.Type) {
	case config.TxFilterType_None:
		return c, nil
	case config.TxFilterType_BirdsNest:
		err := CheckBNConfig(conf.BirdsNest, true)
		if err != StringNil {
			return nil, errors.New(err)
		}
		cuckoo := &common.CuckooConfig{
			KeyType:       common.KeyType(conf.BirdsNest.Cuckoo.KeyType),
			TagsPerBucket: conf.BirdsNest.Cuckoo.TagsPerBucket,
			BitsPerItem:   conf.BirdsNest.Cuckoo.BitsPerItem,
			MaxNumKeys:    conf.BirdsNest.Cuckoo.MaxNumKeys,
			TableType:     conf.BirdsNest.Cuckoo.TableType,
		}
		rules := &common.RulesConfig{
			AbsoluteExpireTime: conf.BirdsNest.Rules.AbsoluteExpireTime,
		}
		snapshot := &common.SnapshotSerializerConfig{
			Type:        common.SerializeIntervalType(conf.BirdsNest.Snapshot.Type),
			Timed:       &common.TimedSerializeIntervalConfig{Interval: int64(conf.BirdsNest.Snapshot.Timed.Interval)},
			BlockHeight: &common.BlockHeightSerializeIntervalConfig{Interval: uint64(conf.BirdsNest.Snapshot.Timed.Interval)},
			Path:        conf.BirdsNest.Snapshot.Path,
		}
		c.BirdsNest = &common.BirdsNestConfig{
			Length:   conf.BirdsNest.Length,
			ChainId:  chainId,
			Rules:    rules,
			Cuckoo:   cuckoo,
			Snapshot: snapshot,
		}
		return c, nil
	case config.TxFilterType_Map:
		return c, nil
	case config.TxFilterType_ShardingBirdsNest:
		err := CheckShardingBNConfig(conf.ShardingBirdsNest)
		if err != StringNil {
			return nil, errors.New(err)
		}
		snapshot := &common.SnapshotSerializerConfig{
			Type:        common.SerializeIntervalType(conf.BirdsNest.Snapshot.Type),
			Timed:       &common.TimedSerializeIntervalConfig{Interval: int64(conf.BirdsNest.Snapshot.Timed.Interval)},
			BlockHeight: &common.BlockHeightSerializeIntervalConfig{Interval: uint64(conf.BirdsNest.Snapshot.Timed.Interval)},
			Path:        conf.ShardingBirdsNest.Snapshot.Path,
		}
		rules := &common.RulesConfig{
			AbsoluteExpireTime: conf.ShardingBirdsNest.BirdsNest.Rules.AbsoluteExpireTime,
		}
		cuckoo := &common.CuckooConfig{
			KeyType:       common.KeyType(conf.ShardingBirdsNest.BirdsNest.Cuckoo.KeyType),
			TagsPerBucket: conf.ShardingBirdsNest.BirdsNest.Cuckoo.TagsPerBucket,
			BitsPerItem:   conf.ShardingBirdsNest.BirdsNest.Cuckoo.BitsPerItem,
			MaxNumKeys:    conf.ShardingBirdsNest.BirdsNest.Cuckoo.MaxNumKeys,
			TableType:     conf.ShardingBirdsNest.BirdsNest.Cuckoo.TableType,
		}
		bn := &common.BirdsNestConfig{
			Length:   conf.ShardingBirdsNest.BirdsNest.Length,
			ChainId:  chainId,
			Rules:    rules,
			Cuckoo:   cuckoo,
			Snapshot: proto.Clone(snapshot).(*common.SnapshotSerializerConfig),
		}
		c.ShardingBirdsNest = &common.ShardingBirdsNestConfig{
			ChainId:   chainId,
			Length:    conf.ShardingBirdsNest.Length,
			Timeout:   conf.ShardingBirdsNest.Timeout,
			Birdsnest: bn,
			Snapshot:  snapshot,
		}
		return c, nil
	default:
		return c, nil
	}
}

//func ToLocalhost(conf config.TxFilterConfig) localconf.TxFilterConfig {
//	return localconf.TxFilterConfig{}
//}

type TxFilterLogger struct {
	log protocol.Logger
}

func (t TxFilterLogger) Debugf(format string, args ...interface{}) {
	t.log.DebugDynamic(LoggingFixLengthFunc(format, args...))
}

func (t TxFilterLogger) Errorf(format string, args ...interface{}) {
	t.log.Errorf(format, args...)
}

func (t TxFilterLogger) Infof(format string, args ...interface{}) {
	t.log.Infof(format, args...)
}

func NewLogger(log protocol.Logger) *TxFilterLogger {
	return &TxFilterLogger{log: log}
}

func LoggingFixLengthFunc(format string, args ...interface{}) func() string {
	return func() string {
		return LoggingFixLength(format, args...)
	}
}
func LoggingFixLength(format string, args ...interface{}) string {
	str := fmt.Sprintf(format, args...)
	if len(str) > 1024 {
		str = str[:1024] + "..."
	}
	return str
}
