/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package defau1t

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

type TxFilter struct {
	store protocol.BlockchainStore
}

func Init(store protocol.BlockchainStore) *TxFilter {
	return &TxFilter{store: store}
}

func (f TxFilter) IsExistsAndReturnHeight(_ string, _ ...common.RuleType) (bool, uint64, error) {
	return false, 0, nil
}

func (f TxFilter) SetHeight(_ uint64) {
}

func (f TxFilter) GetHeight() uint64 {
	return 0
}

func (f TxFilter) Add(_ string) error {
	return nil
}

func (f TxFilter) Adds(_ []string) error {
	return nil
}

func (f TxFilter) IsExists(txId string, _ ...common.RuleType) (bool, error) {
	return f.store.TxExists(txId)
}

func (f TxFilter) Close() {
}
