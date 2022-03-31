/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package mapimpl

import (
	"reflect"
	"sync"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/common"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		want *TxFilter
	}{
		{
			name: "test0",
			want: &TxFilter{m: sync.Map{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_IsExistsAndReturnHeight(t *testing.T) {
	type fields struct {
		height uint64
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
			fields: fields{
				height: 0,
			},
			args: args{
				in0: "",
				in1: nil,
			},
			want:    false,
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
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

func TestTxFilter_GetHeight(t *testing.T) {
	type fields struct {
		height uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			name: "test0",
			fields: fields{
				height: 0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
			}
			if got := f.GetHeight(); got != tt.want {
				t.Errorf("GetHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxFilter_SetHeight(t *testing.T) {
	type fields struct {
		height uint64
	}
	type args struct {
		height uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				height: 0,
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
			}
			f.SetHeight(tt.args.height)
		})
	}
}

func TestTxFilter_Add(t *testing.T) {
	type fields struct {
		height uint64
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
			name: "test0",
			fields: fields{
				height: 0,
			},
			args: args{
				txId: "chain1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
				m:      sync.Map{},
			}
			if err := f.Add(tt.args.txId); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_Adds(t *testing.T) {
	type fields struct {
		height uint64
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
			name:    "test0",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				height: 0,
			},
			args: args{
				txIds: []string{"1", "2", "3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
				m:      sync.Map{},
			}
			if err := f.Adds(tt.args.txIds); (err != nil) != tt.wantErr {
				t.Errorf("Adds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxFilter_IsExists(t *testing.T) {
	type fields struct {
		height uint64
	}
	type args struct {
		txId string
		in1  []common.RuleType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "test0",
			fields:  fields{},
			args:    args{},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
				m:      sync.Map{},
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
		height uint64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				height: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TxFilter{
				height: tt.fields.height,
			}
			f.Close()
		})
	}
}
