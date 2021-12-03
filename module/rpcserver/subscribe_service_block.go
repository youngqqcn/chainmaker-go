/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package rpcserver

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker-go/module/subscriber"
	"chainmaker.org/chainmaker-go/module/subscriber/model"
	"chainmaker.org/chainmaker/common/v2/bytehelper"
	commonErr "chainmaker.org/chainmaker/common/v2/errors"
	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	storePb "chainmaker.org/chainmaker/pb-go/v2/store"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	protocol "chainmaker.org/chainmaker/protocol/v2"
	utils "chainmaker.org/chainmaker/utils/v2"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// dealBlockSubscription - deal block subscribe request
func (s *ApiService) dealBlockSubscription(tx *commonPb.Transaction, server apiPb.RpcNode_SubscribeServer) error {
	var (
		err             error
		errMsg          string
		errCode         commonErr.ErrCode
		db              protocol.BlockchainStore
		lastBlockHeight int64
		payload         = tx.Payload
		startBlock      int64
		endBlock        int64
		withRWSet       = false
		onlyHeader      = false
		reqSender       protocol.Role
	)

	for _, kv := range payload.Parameters {
		if kv.Key == syscontract.SubscribeBlock_START_BLOCK.String() {
			startBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeBlock_END_BLOCK.String() {
			endBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeBlock_WITH_RWSET.String() {
			if string(kv.Value) == TRUE {
				withRWSet = true
			}
		} else if kv.Key == syscontract.SubscribeBlock_ONLY_HEADER.String() {
			if string(kv.Value) == TRUE {
				onlyHeader = true
				withRWSet = false
			}
		}

		if err != nil {
			errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_BLOCK
			errMsg = s.getErrMsg(errCode, err)
			s.log.Error(errMsg)
			return status.Error(codes.InvalidArgument, errMsg)
		}
	}

	if err = s.checkSubscribeBlockHeight(startBlock, endBlock); err != nil {
		errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_BLOCK
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	s.log.Infof("Recv block subscribe request: [start:%d]/[end:%d]/[withRWSet:%v]/[onlyHeader:%v]",
		startBlock, endBlock, withRWSet, onlyHeader)

	chainId := tx.Payload.ChainId
	if db, err = s.chainMakerServer.GetStore(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_STORE
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	if lastBlockHeight, err = s.checkAndGetLastBlockHeight(db, startBlock); err != nil {
		errCode = commonErr.ERR_CODE_GET_LAST_BLOCK
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	reqSender, err = s.getRoleFromTx(tx)
	reqSenderOrgId := tx.Sender.Signer.OrgId
	if err != nil {
		return err
	}

	var startBlockHeight int64
	if startBlock > startBlockHeight {
		startBlockHeight = startBlock
	}

	if startBlock == -1 && endBlock == -1 {
		return s.sendNewBlock(db, tx, server, endBlock, withRWSet, onlyHeader,
			-1, reqSender, reqSenderOrgId)
	}

	if endBlock != -1 && endBlock <= lastBlockHeight {
		_, err = s.sendHistoryBlock(db, server, startBlockHeight, endBlock,
			withRWSet, onlyHeader, reqSender, reqSenderOrgId)

		if err != nil {
			s.log.Errorf("sendHistoryBlock failed, %s", err)
			return err
		}

		return status.Error(codes.OK, "OK")
	}

	alreadySendHistoryBlockHeight, err := s.sendHistoryBlock(db, server, startBlockHeight, endBlock,
		withRWSet, onlyHeader, reqSender, reqSenderOrgId)

	if err != nil {
		s.log.Errorf("sendHistoryBlock failed, %s", err)
		return err
	}

	s.log.Debugf("after sendHistoryBlock, alreadySendHistoryBlockHeight is %d", alreadySendHistoryBlockHeight)

	return s.sendNewBlock(db, tx, server, endBlock, withRWSet, onlyHeader, alreadySendHistoryBlockHeight,
		reqSender, reqSenderOrgId)
}

// sendNewBlock - send new block to subscriber
func (s *ApiService) sendNewBlock(store protocol.BlockchainStore, tx *commonPb.Transaction,
	server apiPb.RpcNode_SubscribeServer,
	endBlockHeight int64, withRWSet, onlyHeader bool, alreadySendHistoryBlockHeight int64,
	reqSender protocol.Role, reqSenderOrgId string) error {

	var (
		errCode         commonErr.ErrCode
		err             error
		errMsg          string
		eventSubscriber *subscriber.EventSubscriber
		blockInfo       *commonPb.BlockInfo
	)

	blockCh := make(chan model.NewBlockEvent)

	chainId := tx.Payload.ChainId
	if eventSubscriber, err = s.chainMakerServer.GetEventSubscribe(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_SUBSCRIBER
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	sub := eventSubscriber.SubscribeBlockEvent(blockCh)
	defer sub.Unsubscribe()

	for {
		select {
		case ev := <-blockCh:
			blockInfo = ev.BlockInfo

			if alreadySendHistoryBlockHeight != -1 && int64(blockInfo.Block.Header.BlockHeight) > alreadySendHistoryBlockHeight {
				_, err = s.sendHistoryBlock(store, server, alreadySendHistoryBlockHeight+1,
					int64(blockInfo.Block.Header.BlockHeight), withRWSet, onlyHeader, reqSender, reqSenderOrgId)
				if err != nil {
					s.log.Errorf("send history block failed, %s", err)
					return err
				}

				alreadySendHistoryBlockHeight = -1
				continue
			}

			if reqSender == protocol.RoleLight {
				newBlock := utils.FilterBlockTxs(reqSenderOrgId, blockInfo.Block)
				blockInfo = &commonPb.BlockInfo{
					Block:     newBlock,
					RwsetList: ev.BlockInfo.RwsetList,
				}
			}

			//printAllTxsOfBlock(blockInfo, reqSender, reqSenderOrgId)

			if err = s.dealBlockSubscribeResult(server, blockInfo, withRWSet, onlyHeader); err != nil {
				s.log.Errorf(err.Error())
				return status.Error(codes.Internal, err.Error())
			}

			if endBlockHeight != -1 && int64(blockInfo.Block.Header.BlockHeight) >= endBlockHeight {
				return status.Error(codes.OK, "OK")
			}

		case <-server.Context().Done():
			return nil
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *ApiService) dealBlockSubscribeResult(server apiPb.RpcNode_SubscribeServer, blockInfo *commonPb.BlockInfo,
	withRWSet, onlyHeader bool) error {

	var (
		err    error
		result *commonPb.SubscribeResult
	)

	if !withRWSet {
		blockInfo = &commonPb.BlockInfo{
			Block:     blockInfo.Block,
			RwsetList: nil,
		}
	}

	if result, err = s.getBlockSubscribeResult(blockInfo, onlyHeader); err != nil {
		return fmt.Errorf("get block subscribe result failed, %s", err)
	}

	if err := server.Send(result); err != nil {
		return fmt.Errorf("send block subscribe result by realtime failed, %s", err)
	}

	return nil
}

// sendHistoryBlock - send history block to subscriber
func (s *ApiService) sendHistoryBlock(store protocol.BlockchainStore, server apiPb.RpcNode_SubscribeServer,
	startBlockHeight, endBlockHeight int64, withRWSet, onlyHeader bool, reqSender protocol.Role,
	reqSenderOrgId string) (int64, error) {

	var (
		err    error
		errMsg string
		result *commonPb.SubscribeResult
	)

	i := startBlockHeight
	for {
		select {
		case <-s.ctx.Done():
			return -1, status.Error(codes.Internal, "chainmaker is restarting, please retry later")
		default:
			if err = s.getRateLimitToken(); err != nil {
				return -1, status.Error(codes.Internal, err.Error())
			}

			if endBlockHeight != -1 && i > endBlockHeight {
				return i - 1, nil
			}

			blockInfo, alreadySendHistoryBlockHeight, err := s.getBlockInfoFromStore(store, i, withRWSet,
				reqSender, reqSenderOrgId)

			if err != nil {
				return -1, status.Error(codes.Internal, errMsg)
			}

			if blockInfo == nil || alreadySendHistoryBlockHeight > 0 {
				return alreadySendHistoryBlockHeight, nil
			}

			if result, err = s.getBlockSubscribeResult(blockInfo, onlyHeader); err != nil {
				errMsg = fmt.Sprintf("get block subscribe result failed, %s", err)
				s.log.Error(errMsg)
				return -1, errors.New(errMsg)
			}

			if err := server.Send(result); err != nil {
				errMsg = fmt.Sprintf("send block info by history failed, %s", err)
				s.log.Error(errMsg)
				return -1, status.Error(codes.Internal, errMsg)
			}

			i++
		}
	}
}

func (s *ApiService) getBlockSubscribeResult(blockInfo *commonPb.BlockInfo,
	onlyHeader bool) (*commonPb.SubscribeResult, error) {

	var (
		resultBytes []byte
		err         error
	)

	if onlyHeader {
		resultBytes, err = proto.Marshal(blockInfo.Block.Header)
	} else {
		resultBytes, err = proto.Marshal(blockInfo)
	}

	if err != nil {
		errMsg := fmt.Sprintf("marshal block subscribe result failed, %s", err)
		s.log.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	result := &commonPb.SubscribeResult{
		Data: resultBytes,
	}

	return result, nil
}

func (s *ApiService) getBlockInfoFromStore(store protocol.BlockchainStore, curblockHeight int64, withRWSet bool,
	reqSender protocol.Role, reqSenderOrgId string) (blockInfo *commonPb.BlockInfo,
	alreadySendHistoryBlockHeight int64, err error) {

	var (
		errMsg         string
		block          *commonPb.Block
		blockWithRWSet *storePb.BlockWithRWSet
	)

	if withRWSet {
		blockWithRWSet, err = store.GetBlockWithRWSets(uint64(curblockHeight))
	} else {
		block, err = store.GetBlock(uint64(curblockHeight))
	}

	if err != nil {
		if withRWSet {
			errMsg = fmt.Sprintf("get block with rwset failed, at [height:%d], %s", curblockHeight, err)
		} else {
			errMsg = fmt.Sprintf("get block failed, at [height:%d], %s", curblockHeight, err)
		}
		s.log.Error(errMsg)
		return nil, -1, errors.New(errMsg)
	}

	if withRWSet {
		if blockWithRWSet == nil {
			return nil, curblockHeight - 1, nil
		}

		blockInfo = &commonPb.BlockInfo{
			Block:     blockWithRWSet.Block,
			RwsetList: blockWithRWSet.TxRWSets,
		}

		// filter txs so that only related ones get passed
		if reqSender == protocol.RoleLight {
			newBlock := utils.FilterBlockTxs(reqSenderOrgId, blockWithRWSet.Block)
			blockInfo = &commonPb.BlockInfo{
				Block:     newBlock,
				RwsetList: blockWithRWSet.TxRWSets,
			}
		}
	} else {
		if block == nil {
			return nil, curblockHeight - 1, nil
		}

		blockInfo = &commonPb.BlockInfo{
			Block:     block,
			RwsetList: nil,
		}

		// filter txs so that only related ones get passed
		if reqSender == protocol.RoleLight {
			newBlock := utils.FilterBlockTxs(reqSenderOrgId, block)
			blockInfo = &commonPb.BlockInfo{
				Block:     newBlock,
				RwsetList: nil,
			}
		}
	}

	//printAllTxsOfBlock(blockInfo, reqSender, reqSenderOrgId)

	return blockInfo, -1, nil
}

//func printAllTxsOfBlock(blockInfo *commonPb.BlockInfo, reqSender protocol.Role, reqSenderOrgId string) {
//	fmt.Printf("Verifying subscribed block of height: %d\n", blockInfo.Block.Header.BlockHeight)
//	fmt.Printf("verify: the role of request sender is Light [%t]\n", reqSender == protocol.RoleLight)
//	fmt.Printf("the block has %d txs\n", len(blockInfo.Block.Txs))
//	for i, tx := range blockInfo.Block.Txs {
//
//		if tx.Sender != nil {
//
//			fmt.Printf("Tx [%d] of subscribed block, from org %v, TxSenderOrgId is %s, "+
//				"verify: this tx is of the same organization [%t]\n", i, tx.Sender.Signer.OrgId,
//				reqSenderOrgId, tx.Sender.Signer.OrgId == reqSenderOrgId)
//		}
//	}
//	fmt.Println()
//}
