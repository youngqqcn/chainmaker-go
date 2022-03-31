/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
)

type ManagerImpl struct {
	snapshots map[utils.BlockFingerPrint]*SnapshotImpl
	delegate  *ManagerDelegate
	log       protocol.Logger
}

func (m *ManagerImpl) storeAndLinkSnapshotImpl(snapshotImpl *SnapshotImpl,
	prevFingerPrint *utils.BlockFingerPrint, fingerPrint *utils.BlockFingerPrint) {
	// 存储当前指纹的snapshot
	m.snapshots[*fingerPrint] = snapshotImpl

	// 如果前序指纹对应的snapshot存在, 就建立snapshot的对应关系
	if prevSnapshot, ok := m.snapshots[*prevFingerPrint]; ok {
		snapshotImpl.SetPreSnapshot(prevSnapshot)
	}
}

// When generating blocks, generate a Snapshot for each block, which is used as read-write set cache
func (m *ManagerImpl) NewSnapshot(prevBlock *commonPb.Block, block *commonPb.Block) protocol.Snapshot {
	m.delegate.lock.Lock()
	defer m.delegate.lock.Unlock()
	blockHeight := block.Header.BlockHeight
	snapshotImpl := m.delegate.makeSnapshotImpl(block)

	// 计算前序指纹, 和当前指纹
	prevFingerPrint := utils.CalcBlockFingerPrintWithoutTx(prevBlock)
	fingerPrint := utils.CalcBlockFingerPrintWithoutTx(block)
	m.storeAndLinkSnapshotImpl(snapshotImpl, &prevFingerPrint, &fingerPrint)

	m.log.Infof(
		"create snapshot@%s at height %d, fingerPrint[%v] -> prevFingerPrint[%v]",
		block.Header.ChainId,
		blockHeight,
		fingerPrint,
		prevFingerPrint,
	)
	return snapshotImpl
}

func (m *ManagerImpl) NotifyBlockCommitted(block *commonPb.Block) error {
	m.delegate.lock.Lock()
	defer m.delegate.lock.Unlock()

	m.log.Infof("commit snapshot@%s at height %d", block.Header.ChainId, block.Header.BlockHeight)

	// 计算刚落块的区块指纹
	deleteFp := utils.CalcBlockFingerPrintWithoutTx(block)
	deleteFpEx := calcNotConsensusFingerPrint(block)
	// 如果有snapshot对应的前序snapshot的指纹, 等于刚落块的区块指纹
	for _, snapshot := range m.snapshots {
		if snapshot == nil || snapshot.GetPreSnapshot() == nil {
			continue
		}
		prevFp := m.delegate.calcSnapshotFingerPrintWithoutTx(snapshot.GetPreSnapshot().(*SnapshotImpl))
		if deleteFp == prevFp {
			snapshot.SetPreSnapshot(nil)
		}
	}

	delete(m.snapshots, deleteFp)

	// 删除未共识的区块指纹
	if _, ok := m.snapshots[deleteFpEx]; ok {
		delete(m.snapshots, deleteFpEx)
		m.log.Infof("delete snapshot@%s %v & %v at height %d",
			block.Header.ChainId, deleteFp, deleteFpEx, block.Header.BlockHeight)
	} else {
		m.log.Infof("delete snapshot@%s %v at height %d",
			block.Header.ChainId, deleteFp, block.Header.BlockHeight)
	}

	snapshotCount := 0
	// in case of switch-fork, gc too old snapshot
	for finger, snapshot := range m.snapshots {
		snapshotCount++
		if snapshot == nil {
			continue
		}

		if block.Header.BlockHeight-snapshot.GetBlockHeight() > 8 {
			delete(m.snapshots, finger)
			m.log.Infof("delete snapshot %v at height %d while gc", finger, snapshot.blockHeight)
			snapshot.SetPreSnapshot(nil)
		}
	}
	m.log.Debugf("current snapshot count:%d", snapshotCount)
	return nil
}

func calcNotConsensusFingerPrint(block *commonPb.Block) utils.BlockFingerPrint {
	if block == nil {
		return ""
	}

	newBlock := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:        block.Header.ChainId,
			BlockHeight:    block.Header.BlockHeight,
			PreBlockHash:   block.Header.PreBlockHash,
			BlockTimestamp: block.Header.BlockTimestamp,
			Proposer:       block.Header.Proposer,
		},
	}

	return utils.CalcBlockFingerPrintWithoutTx(newBlock)
}
