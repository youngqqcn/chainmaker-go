/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package snapshot

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
)

func TestNewSnapshotManager(t *testing.T) {
	t.Log("TestNewSnapshotManager")
	var (
		snapshotFactory Factory
		log             = logger.GetLogger(logger.MODULE_SNAPSHOT)
		ctl             = gomock.NewController(t)
		store           = mock.NewMockBlockchainStore(ctl)
	)

	manager := snapshotFactory.NewSnapshotManager(store, log)

	fmt.Println(manager)
	log.Debug("test NewSnapshotManager")
}

func TestNewSnapshotEvidenceMgr(t *testing.T) {
	t.Log("TestNewSnapshotEvidenceMgr")
	var (
		snapshotFactory Factory
		log             = logger.GetLogger(logger.MODULE_SNAPSHOT)
		ctl             = gomock.NewController(t)
		store           = mock.NewMockBlockchainStore(ctl)
	)

	manager := snapshotFactory.NewSnapshotManager(store, log)

	fmt.Println(manager)
	log.Debug("test NewSnapshotEvidenceMgr")
}
