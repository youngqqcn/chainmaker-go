/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package mapimpl

import (
	"sync"

	"chainmaker.org/chainmaker/pb-go/v2/common"
)

type TxFilter struct {
	height uint64
	m      sync.Map
}

func Init() *TxFilter {
	return &TxFilter{m: sync.Map{}}
}

func (f *TxFilter) IsExistsAndReturnHeight(_ string, _ ...common.RuleType) (bool, uint64, error) {
	return false, 0, nil
}

func (f *TxFilter) GetHeight() uint64 {
	return 0
}

func (f *TxFilter) SetHeight(height uint64) {
	f.height = height
}

func (f *TxFilter) Add(txId string) error {
	f.m.Store(txId, struct{}{})
	return nil
}

func (f *TxFilter) Adds(txIds []string) error {
	for _, txId := range txIds {
		f.m.Store(txId, struct{}{})
	}
	return nil
}

func (f *TxFilter) IsExists(txId string, _ ...common.RuleType) (bool, error) {
	_, ok := f.m.Load(txId)
	return ok, nil
}

func (f *TxFilter) Close() {
}
