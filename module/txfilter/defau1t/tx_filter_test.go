/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package defau1t

import (
	"reflect"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/golang/mock/gomock"
)

func TestInit(t *testing.T) {
	type args struct {
		store protocol.BlockchainStore
	}

	store := newMockBlockchainStore(t)
	tests := []struct {
		name string
		args args
		want *TxFilter
	}{
		{
			name: "test0",
			args: args{
				store: store,
			},
			want: &TxFilter{
				store: store,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.store); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_IsExistsAndReturnHeight(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	type args struct {
		in0 string
		in1 []common.RuleType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   uint64
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{store: func() protocol.BlockchainStore {
				store := mock.NewMockBlockchainStore(gomock.NewController(t))
				store.EXPECT().TxExistsInFullDB(gomock.Any()).Return(false, uint64(0), nil).AnyTimes()
				return store
			}()},
			args:    args{},
			want:    false,
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			got, got1, err := f.IsExistsAndReturnHeight(tt.args.in0, tt.args.in1...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsExistsAndReturnHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsExistsAndReturnHeight() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsExistsAndReturnHeight() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTxFilter_SetHeight(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	type args struct {
		in0 uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{store: func() protocol.BlockchainStore {
				store := mock.NewMockBlockchainStore(gomock.NewController(t))
				store.EXPECT().TxExistsInFullDB(gomock.Any()).Return(false, uint64(0), nil).AnyTimes()
				return store
			}()},
			args: args{
				uint64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			f.SetHeight(tt.args.in0)
		})
	}
}

func TestTxFilter_GetHeight(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			name: "test0",
			fields: fields{store: func() protocol.BlockchainStore {
				store := mock.NewMockBlockchainStore(gomock.NewController(t))
				store.EXPECT().TxExistsInFullDB(gomock.Any()).Return(false, uint64(0), nil).AnyTimes()
				store.EXPECT().GetLastBlock().Return(&common.Block{Header: &common.BlockHeader{BlockHeight: 0}}, nil).AnyTimes()
				return store
			}()},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			if got := f.GetHeight(); got != tt.want {
				t.Errorf("GetHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_Add(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	type args struct {
		in0 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				store: newMockBlockchainStore(t),
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			if err := f.Add(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_Adds(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	type args struct {
		in0 []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				store: newMockBlockchainStore(t),
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			if err := f.Adds(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("Adds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_IsExists(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}
	type args struct {
		txId string
		in1  []common.RuleType
	}

	var (
		store = newMockBlockchainStore(t)
	)
	store.EXPECT().TxExists(gomock.Any()).AnyTimes()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				store: store,
			},
			args:    args{},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			got, err := f.IsExists(tt.args.txId, tt.args.in1...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_Close(t *testing.T) {
	type fields struct {
		store protocol.BlockchainStore
	}

	store := newMockBlockchainStore(t)
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				store: store,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TxFilter{
				store: tt.fields.store,
			}
			f.Close()
		})
	}
}

func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}
