/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBlockSyncServerConf(t *testing.T) {
	conf := NewBlockSyncServerConf()
	conf.SetBlockPoolSize(10)
	require.Equal(t, conf.blockPoolSize, uint64(10))
	conf.SetWaitTimeOfBlockRequestMsg(10)
	require.Equal(t, conf.timeOut, 10*time.Second)
	conf.SetBatchSizeFromOneNode(10)
	require.Equal(t, conf.batchSizeFromOneNode, uint64(10))
	conf.SetProcessBlockTicker(10)
	require.Equal(t, conf.processBlockTick, 10*time.Second)
	conf.SetSchedulerTicker(10)
	require.Equal(t, conf.schedulerTick, 10*time.Second)
	conf.SetLivenessTicker(10)
	require.Equal(t, conf.livenessTick, 10*time.Second)
	conf.SetNodeStatusTicker(10)
	require.Equal(t, conf.nodeStatusTick, 10*time.Second)
	conf.SetDataDetectionTicker(10)
	require.Equal(t, conf.dataDetectionTick, 10*time.Second)
	conf.SetReqTimeThreshold(10)
	require.Equal(t, conf.reqTimeThreshold, 10*time.Second)
	conf.SetBlockRequestTime(10)
	require.Equal(t, conf.blockRequestTime, 10*time.Second)
}
