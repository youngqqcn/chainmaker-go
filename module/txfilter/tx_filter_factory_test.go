/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package txfilter

import (
	"reflect"
	"strconv"
	"testing"

	"chainmaker.org/chainmaker-go/module/txfilter/birdnest"
	"chainmaker.org/chainmaker-go/module/txfilter/defau1t"
	mapimpl "chainmaker.org/chainmaker-go/module/txfilter/map"
	"chainmaker.org/chainmaker-go/module/txfilter/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/golang/mock/gomock"
)

func Test_txFilterFactory_NewTxFilter(t *testing.T) {
	type args struct {
		conf  *config.TxFilterConfig
		log   protocol.Logger
		store protocol.BlockchainStore
	}

	var (
		log   = newMockLogger(t)
		store = newMockBlockchainStore(t)
		block = createBlockByHash(0, []byte("123456"))
	)

	store.EXPECT().GetLastBlock().Return(block, nil).AnyTimes()
	log.EXPECT().DebugDynamic(gomock.Any()).AnyTimes()
	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	defaultConf := GetTestDefaultConfig("test", 0)

	log.EXPECT().Warn(gomock.Any()).AnyTimes()
	tests := []struct {
		name    string
		args    args
		want    protocol.TxFilter
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				conf:  nil,
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return defau1t.New(store)
			}(),
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				conf: &config.TxFilterConfig{
					Type:      config.TxFilterType_None,
					BirdsNest: defaultConf.Birdsnest,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return defau1t.New(store)
			}(),
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				conf: &config.TxFilterConfig{
					Type:      config.TxFilterType_BirdsNest,
					BirdsNest: defaultConf.Birdsnest,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter, err := birdnest.New(defaultConf.Birdsnest, log, store)
				if err != nil {
					t.Log(err)
					return nil
				}
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				conf: &config.TxFilterConfig{
					Type:              config.TxFilterType_Map,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter := mapimpl.New()
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				conf: &config.TxFilterConfig{
					Type:              config.TxFilterType_ShardingBirdsNest,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				txFilter, err := shardingbirdsnest.New(defaultConf, log, store)
				if err != nil {
					t.Log(err)
					return nil
				}
				return txFilter
			}(),
			wantErr: false,
		},
		{
			name: "test5",
			args: args{
				conf: &config.TxFilterConfig{
					Type:              5,
					BirdsNest:         defaultConf.Birdsnest,
					ShardingBirdsNest: defaultConf,
				},
				log:   log,
				store: store,
			},
			want: func() protocol.TxFilter {
				return defau1t.New(store)
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &txFilterFactory{}
			got, err := cf.NewTxFilter(tt.args.conf, tt.args.log, tt.args.store)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTxFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.name == "test2" || tt.name == "test4" {
				if !reflect.DeepEqual(got.GetHeight(), tt.want.GetHeight()) {
					t.Errorf("NewTxFilter() got = %v, want %v", got, tt.want)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewTxFilter() got = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}

func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

func createBlockByHash(height uint64, hash []byte) *commonpb.Block {
	//var hash = []byte("0123456789")
	var version = uint32(1)
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			ChainId:        "Chain1",
			BlockHeight:    height,
			PreBlockHash:   hash,
			BlockHash:      hash,
			PreConfHeight:  0,
			BlockVersion:   version,
			DagHash:        hash,
			RwSetRoot:      hash,
			TxRoot:         hash,
			BlockTimestamp: 0,
			Proposer:       &accesscontrol.Member{MemberInfo: hash},
			ConsensusArgs:  nil,
			TxCount:        1,
			Signature:      []byte(""),
		},
		Dag: &commonpb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}

	return block
}

func GetTestDefaultConfig(path string, i int) *commonpb.ShardingBirdsNestConfig {
	return &commonpb.ShardingBirdsNestConfig{
		Length:  10,
		Timeout: 10,
		ChainId: "chain1",
		Birdsnest: &commonpb.BirdsNestConfig{
			ChainId: "chain1",
			Length:  5,
			Rules: &commonpb.RulesConfig{
				AbsoluteExpireTime: 10000,
			},
			Cuckoo: &commonpb.CuckooConfig{
				KeyType:       commonpb.KeyType_KTDefault,
				TagsPerBucket: 4,
				BitsPerItem:   9,
				MaxNumKeys:    10,
				TableType:     1,
			},
			Snapshot: &commonpb.SnapshotSerializerConfig{
				Type:        commonpb.SerializeIntervalType_Timed,
				Timed:       &commonpb.TimedSerializeIntervalConfig{Interval: 20},
				BlockHeight: &commonpb.BlockHeightSerializeIntervalConfig{Interval: 20},
				Path:        path + strconv.Itoa(i),
			},
		},
		Snapshot: &commonpb.SnapshotSerializerConfig{
			Type:        commonpb.SerializeIntervalType_Timed,
			Timed:       &commonpb.TimedSerializeIntervalConfig{Interval: 20},
			BlockHeight: &commonpb.BlockHeightSerializeIntervalConfig{Interval: 20},
			Path:        path + strconv.Itoa(i),
		},
	}
}
