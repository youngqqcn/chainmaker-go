/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package defau1t

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// TxFilter protocol.BlockchainStore transaction filter
type TxFilter struct {
	store protocol.BlockchainStore
}

func (f TxFilter) ValidateRule(_ string, _ ...common.RuleType) error {
	return nil
}

// New transaction filter init
func New(store protocol.BlockchainStore) *TxFilter {
	return &TxFilter{store: store}
}

// GetHeight get height from transaction filter
func (f TxFilter) GetHeight() uint64 {
	block, err := f.store.GetLastBlock()
	if err != nil {
		return 0
	}
	return block.Header.BlockHeight
}

// SetHeight set height from transaction filter
func (f TxFilter) SetHeight(_ uint64) {
}

// IsExistsAndReturnHeight is exists and return height
func (f TxFilter) IsExistsAndReturnHeight(txId string, _ ...common.RuleType) (bool, uint64, error) {
	return f.store.TxExistsInFullDB(txId)
}

// Add txId to transaction filter
func (f TxFilter) Add(_ string) error {
	return nil
}

// Adds batch Add txId
func (f TxFilter) Adds(_ []string) error {
	return nil
}

// AddsAndSetHeight batch add tx id and set height
func (f TxFilter) AddsAndSetHeight(_ []string, _ uint64) error {
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f TxFilter) IsExists(txId string, _ ...common.RuleType) (bool, error) {
	return f.store.TxExists(txId)
}

// Close transaction filter
func (f TxFilter) Close() {
}
