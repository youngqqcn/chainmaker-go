/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package shardingbirdsnest

import (
	"fmt"
	"strconv"
	"testing"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/golang/mock/gomock"
)

func TestTxFilter_Add(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		log   protocol.Logger
		bn    *sbn.ShardingBirdsNest
		store protocol.BlockchainStore
		exitC chan struct{}
	}
	type args struct {
		txId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "正常流 TimestampKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: utils.GetTimestampTxId(),
			},
			wantErr: false,
		},
		{
			name: "正常流 NormalKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: utils.GetRandTxId(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				log:   tt.fields.log,
				bn:    tt.fields.bn,
				store: tt.fields.store,
				exitC: tt.fields.exitC,
			}
			if err := f.Add(tt.args.txId); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_Adds(t *testing.T) {
	ctrl := gomock.NewController(t)
	type fields struct {
		log   protocol.Logger
		bn    *sbn.ShardingBirdsNest
		store protocol.BlockchainStore
		exitC chan struct{}
	}
	type args struct {
		txIds []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "正常流 NormalKey and TimestampKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txIds: func() []string {
					result := make([]string, 0, 100)
					result = append(result, GetNormalKeys(50)...)
					result = append(result, GetTimestampKeys(50)...)
					return result
				}(),
			},
			wantErr: false,
		},
		{
			name: "正常流 NormalKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txIds: GetNormalKeys(50),
			},
			wantErr: false,
		},
		{
			name: "正常流 TimestampKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txIds: GetTimestampKeys(50),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				log:   tt.fields.log,
				bn:    tt.fields.bn,
				store: tt.fields.store,
				exitC: tt.fields.exitC,
			}
			if err := f.Adds(tt.args.txIds); (err != nil) != tt.wantErr {
				t.Errorf("Adds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	type fields struct {
		log   protocol.Logger
		bn    *sbn.ShardingBirdsNest
		store protocol.BlockchainStore
		exitC chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "正常流",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				log:   tt.fields.log,
				bn:    tt.fields.bn,
				store: tt.fields.store,
				exitC: tt.fields.exitC,
			}
			f.Close()
		})
	}
}

func TestTxFilter_GetHeight(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		log   protocol.Logger
		bn    *sbn.ShardingBirdsNest
		store protocol.BlockchainStore
		exitC chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			name: "正常流",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				log:   tt.fields.log,
				bn:    tt.fields.bn,
				store: tt.fields.store,
				exitC: tt.fields.exitC,
			}
			if got := f.GetHeight(); got != tt.want {
				t.Errorf("GetHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_IsExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	normalKey := utils.GetRandTxId()
	timestampTxId := utils.GetTimestampTxId()
	type fields struct {
		log   protocol.Logger
		bn    *sbn.ShardingBirdsNest
		store protocol.BlockchainStore
		exitC chan struct{}
	}
	type args struct {
		txId     string
		ruleType []common.RuleType
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantExists bool
		wantErr    bool
	}{
		{
			name: "正常流 normalKey",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: func() *mock.MockBlockchainStore {
					store := mock.NewMockBlockchainStore(ctrl)
					store.EXPECT().TxExists(normalKey).Return(true, nil).AnyTimes()
					return store
				}(),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: normalKey,
			},
			wantErr:    false,
			wantExists: true,
		},
		{
			name: "正常流 timestampTxId 过滤器存在 数据库存在",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					key, err := bn.ToTimestampKey(timestampTxId)
					if err != nil {
						t.Error(err)
						return nil
					}
					_ = nest.Add(key)
					return nest
				}(),
				store: func() *mock.MockBlockchainStore {
					store := mock.NewMockBlockchainStore(ctrl)
					store.EXPECT().TxExists(timestampTxId).Return(true, nil).AnyTimes()
					return store
				}(),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: timestampTxId,
			},
			wantErr:    false,
			wantExists: true,
		},
		{
			name: "正常流 timestampTxId 过滤器存在 数据库不存在",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					key, err := bn.ToTimestampKey(timestampTxId)
					if err != nil {
						t.Error(err)
						return nil
					}
					_ = nest.Add(key)
					return nest
				}(),
				store: func() *mock.MockBlockchainStore {
					store := mock.NewMockBlockchainStore(ctrl)
					store.EXPECT().TxExists(timestampTxId).Return(false, nil).AnyTimes()
					return store
				}(),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: timestampTxId,
			},
			wantErr:    false,
			wantExists: false,
		},
		{
			name: "正常流 timestampTxId 过滤器不存在",
			fields: fields{
				log: TestLogger{T: t},
				bn: func() *sbn.ShardingBirdsNest {
					nest, err := sbn.NewShardingBirdsNest(GetTestDefaultConfig(TestDir, 1),
						make(chan struct{}),
						bn.LruStrategy,
						sbn.NewModuloSA(10),
						bn.TestLogger{T: t},
					)
					if err != nil {
						t.Errorf("init error %v", err)
						return nil
					}
					return nest
				}(),
				store: mock.NewMockBlockchainStore(ctrl),
				exitC: make(chan struct{}),
			},
			args: args{
				txId: timestampTxId,
			},
			wantErr:    false,
			wantExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				log:   tt.fields.log,
				bn:    tt.fields.bn,
				store: tt.fields.store,
				exitC: tt.fields.exitC,
			}
			gotExists, err := f.IsExists(tt.args.txId, tt.args.ruleType...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("IsExists() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

const TestDir = "./data/tx_filter"

func GetTestDefaultConfig(path string, i int) *common.ShardingBirdsNestConfig {
	return &common.ShardingBirdsNestConfig{
		Length:  10,
		Timeout: 10,
		ChainId: "chain1",
		Birdsnest: &common.BirdsNestConfig{
			ChainId: "chain1",
			Length:  5,
			Rules: &common.RulesConfig{
				AbsoluteExpireTime: 10000,
			},
			Cuckoo: &common.CuckooConfig{
				KeyType:       common.KeyType_KTDefault,
				TagsPerBucket: 4,
				BitsPerItem:   9,
				MaxNumKeys:    10,
				TableType:     1,
			},
			Snapshot: &common.SnapshotSerializerConfig{
				Type:        common.SerializeIntervalType_Timed,
				Timed:       &common.TimedSerializeIntervalConfig{Interval: 20},
				BlockHeight: &common.BlockHeightSerializeIntervalConfig{Interval: 20},
				Path:        path + strconv.Itoa(i),
			},
		},
		Snapshot: &common.SnapshotSerializerConfig{
			Type:        common.SerializeIntervalType_Timed,
			Timed:       &common.TimedSerializeIntervalConfig{Interval: 20},
			BlockHeight: &common.BlockHeightSerializeIntervalConfig{Interval: 20},
			Path:        path + strconv.Itoa(i),
		},
	}
}

type TestLogger struct {
	T *testing.T
}

func (t TestLogger) Infof(format string, args ...interface{}) {
	t.T.Logf(format, args...)
}

func (t TestLogger) Debug(args ...interface{}) {
	t.T.Log(args...)
}

func (t TestLogger) Debugf(format string, args ...interface{}) {
	t.T.Logf(format, args...)
}

func (t TestLogger) Debugw(msg string, keysAndValues ...interface{}) {
	t.T.Log(fmt.Sprintf(msg, keysAndValues...))
}

func (t TestLogger) Error(args ...interface{}) {
	t.T.Log(args...)
}

func (t TestLogger) Errorf(format string, args ...interface{}) {
	t.T.Errorf(format, args...)
}

func (t TestLogger) Errorw(msg string, keysAndValues ...interface{}) {
	t.T.Errorf(msg, keysAndValues...)
}

func (t TestLogger) Fatal(args ...interface{}) {
	t.T.Fatal(args...)
}

func (t TestLogger) Fatalf(format string, args ...interface{}) {
	t.T.Fatalf(format, args...)
}

func (t TestLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	t.T.Fatalf(msg, keysAndValues...)
}

func (t TestLogger) Info(args ...interface{}) {
	t.T.Log(args...)
}

func (t TestLogger) Infow(msg string, keysAndValues ...interface{}) {
	t.T.Logf(msg, keysAndValues...)
}

func (t TestLogger) Panic(args ...interface{}) {
	t.T.Log(args...)
}

func (t TestLogger) Panicf(format string, args ...interface{}) {
	t.T.Fatalf(format, args...)
}

func (t TestLogger) Panicw(msg string, keysAndValues ...interface{}) {
	t.T.Logf(msg, keysAndValues...)
}

func (t TestLogger) Warn(args ...interface{}) {
	t.T.Log(args...)
}

func (t TestLogger) Warnf(format string, args ...interface{}) {
	t.T.Logf(format, args...)
}

func (t TestLogger) Warnw(msg string, keysAndValues ...interface{}) {
	t.T.Logf(msg, keysAndValues...)
}

func (t TestLogger) DebugDynamic(getStr func() string) {
	t.T.Log(getStr())
}

func (t TestLogger) InfoDynamic(getStr func() string) {
	t.T.Log(getStr())
}

func GetNormalKeys(n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, utils.GetRandTxId())
	}
	return result
}

func GetTimestampKeys(n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, utils.GetRandTxId())
	}
	return result
}
