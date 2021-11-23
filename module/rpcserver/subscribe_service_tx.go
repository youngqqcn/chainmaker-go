/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package rpcserver

import (
	"errors"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker-go/module/subscriber"
	"chainmaker.org/chainmaker-go/module/subscriber/model"
	"chainmaker.org/chainmaker/common/v2/bytehelper"
	commonErr "chainmaker.org/chainmaker/common/v2/errors"
	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	protocol "chainmaker.org/chainmaker/protocol/v2"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// dealTxSubscription - deal tx subscribe request
func (s *ApiService) dealTxSubscription(tx *commonPb.Transaction, server apiPb.RpcNode_SubscribeServer) error {
	var (
		err          error
		errMsg       string
		errCode      commonErr.ErrCode
		db           protocol.BlockchainStore
		payload      = tx.Payload
		startBlock   int64
		endBlock     int64
		contractName string
		txIds        []string
		reqSender    protocol.Role
	)

	for _, kv := range payload.Parameters {
		if kv.Key == syscontract.SubscribeTx_START_BLOCK.String() {
			startBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeTx_END_BLOCK.String() {
			endBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeTx_CONTRACT_NAME.String() {
			contractName = string(kv.Value)
		} else if kv.Key == syscontract.SubscribeTx_TX_IDS.String() {
			if kv.Value != nil {
				txIds = strings.Split(string(kv.Value), ",")
			}
		}

		if err != nil {
			errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_TX
			errMsg = s.getErrMsg(errCode, err)
			s.log.Error(errMsg)
			return status.Error(codes.InvalidArgument, errMsg)
		}
	}

	if err = s.checkSubscribeBlockHeight(startBlock, endBlock); err != nil {
		errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_TX
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	s.log.Infof("Recv tx subscribe request: [start:%d]/[end:%d]/[contractName:%s]/[txIds:%+v]",
		startBlock, endBlock, contractName, txIds)

	chainId := tx.Payload.ChainId
	if db, err = s.chainMakerServer.GetStore(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_STORE
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	reqSender, err = s.getRoleFromTx(tx)
	if err != nil {
		return err
	}
	reqSenderOrgId := tx.Sender.Signer.OrgId
	return s.doSendTx(tx, db, server, startBlock, endBlock, contractName, txIds, reqSender, reqSenderOrgId)
}

func (s *ApiService) doSendTx(tx *commonPb.Transaction, db protocol.BlockchainStore,
	server apiPb.RpcNode_SubscribeServer, startBlock, endBlock int64, contractName string,
	txIds []string, reqSender protocol.Role, reqSenderOrgId string) error {

	var (
		txIdsMap                      = make(map[string]struct{})
		alreadySendHistoryBlockHeight int64
		err                           error
	)

	for _, txId := range txIds {
		txIdsMap[txId] = struct{}{}
	}

	if startBlock == -1 && endBlock == -1 {
		return s.sendNewTx(db, tx, server, startBlock, endBlock, contractName, txIds,
			txIdsMap, -1, reqSender, reqSenderOrgId)
	}

	if alreadySendHistoryBlockHeight, err = s.doSendHistoryTx(db, server, startBlock, endBlock,
		contractName, txIds, txIdsMap, reqSender, reqSenderOrgId); err != nil {
		return err
	}

	if alreadySendHistoryBlockHeight == 0 {
		return status.Error(codes.OK, "OK")
	}

	return s.sendNewTx(db, tx, server, startBlock, endBlock, contractName, txIds, txIdsMap,
		alreadySendHistoryBlockHeight, reqSender, reqSenderOrgId)
}

func (s *ApiService) doSendHistoryTx(db protocol.BlockchainStore, server apiPb.RpcNode_SubscribeServer,
	startBlock, endBlock int64, contractName string, txIds []string,
	txIdsMap map[string]struct{}, reqSender protocol.Role, reqSenderOrgId string) (int64, error) {

	var (
		err             error
		errMsg          string
		errCode         commonErr.ErrCode
		lastBlockHeight int64
	)

	var startBlockHeight int64
	if startBlock > startBlockHeight {
		startBlockHeight = startBlock
	}

	if lastBlockHeight, err = s.checkAndGetLastBlockHeight(db, startBlock); err != nil {
		errCode = commonErr.ERR_CODE_GET_LAST_BLOCK
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return -1, status.Error(codes.Internal, errMsg)
	}

	if endBlock != -1 && endBlock <= lastBlockHeight {
		_, err = s.sendHistoryTx(db, server, startBlockHeight, endBlock, contractName,
			txIds, txIdsMap, reqSender, reqSenderOrgId)

		if err != nil {
			s.log.Errorf("sendHistoryTx failed, %s", err)
			return -1, err
		}

		return 0, status.Error(codes.OK, "OK")
	}

	if len(txIds) > 0 && len(txIdsMap) == 0 {
		return 0, status.Error(codes.OK, "OK")
	}

	alreadySendHistoryBlockHeight, err := s.sendHistoryTx(db, server, startBlockHeight, endBlock, contractName,
		txIds, txIdsMap, reqSender, reqSenderOrgId)

	if err != nil {
		s.log.Errorf("sendHistoryTx failed, %s", err)
		return -1, err
	}

	if len(txIds) > 0 && len(txIdsMap) == 0 {
		return 0, status.Error(codes.OK, "OK")
	}

	s.log.Debugf("after sendHistoryBlock, alreadySendHistoryBlockHeight is %d", alreadySendHistoryBlockHeight)

	return alreadySendHistoryBlockHeight, nil
}

// sendNewTx - send new tx to subscriber
func (s *ApiService) sendNewTx(store protocol.BlockchainStore, tx *commonPb.Transaction,
	server apiPb.RpcNode_SubscribeServer, startBlock, endBlock int64, contractName string,
	txIds []string, txIdsMap map[string]struct{}, alreadySendHistoryBlockHeight int64,
	reqSender protocol.Role, reqSenderOrgId string) error {

	var (
		errCode         commonErr.ErrCode
		err             error
		errMsg          string
		eventSubscriber *subscriber.EventSubscriber
		block           *commonPb.Block
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
			block = ev.BlockInfo.Block

			if alreadySendHistoryBlockHeight != -1 && int64(block.Header.BlockHeight) > alreadySendHistoryBlockHeight {
				_, err = s.sendHistoryTx(store, server, alreadySendHistoryBlockHeight+1,
					int64(block.Header.BlockHeight), contractName, txIds, txIdsMap, reqSender, reqSenderOrgId)
				if err != nil {
					s.log.Errorf("send history block failed, %s", err)
					return err
				}

				alreadySendHistoryBlockHeight = -1
				continue
			}

			if err := s.sendSubscribeTx(server, block.Txs, contractName, txIds, txIdsMap,
				reqSender, reqSenderOrgId); err != nil {
				errMsg = fmt.Sprintf("send subscribe tx failed, %s", err)
				s.log.Error(errMsg)
				return status.Error(codes.Internal, errMsg)
			}

			if s.checkIsFinish(txIds, endBlock, txIdsMap, ev.BlockInfo) {
				return status.Error(codes.OK, "OK")
			}

		case <-server.Context().Done():
			return nil
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *ApiService) checkIsFinish(txIds []string, endBlock int64,
	txIdsMap map[string]struct{}, blockInfo *commonPb.BlockInfo) bool {

	if len(txIds) > 0 && len(txIdsMap) == 0 {
		return true
	}

	if endBlock != -1 && int64(blockInfo.Block.Header.BlockHeight) >= endBlock {
		return true
	}

	return false
}

// sendHistoryTx - send history tx to subscriber
func (s *ApiService) sendHistoryTx(store protocol.BlockchainStore,
	server apiPb.RpcNode_SubscribeServer,
	startBlockHeight, endBlockHeight int64,
	contractName string, txIds []string, txIdsMap map[string]struct{},
	reqSender protocol.Role, reqSenderOrgId string) (int64, error) {

	var (
		err    error
		errMsg string
		block  *commonPb.Block
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

			if len(txIds) > 0 && len(txIdsMap) == 0 {
				return i - 1, nil
			}

			block, err = store.GetBlock(uint64(i))

			if err != nil {
				errMsg = fmt.Sprintf("get block failed, at [height:%d], %s", i, err)
				s.log.Error(errMsg)
				return -1, status.Error(codes.Internal, errMsg)
			}

			if block == nil {
				return i - 1, nil
			}

			if err := s.sendSubscribeTx(server, block.Txs, contractName, txIds, txIdsMap,
				reqSender, reqSenderOrgId); err != nil {
				errMsg = fmt.Sprintf("send subscribe tx failed, %s", err)
				s.log.Error(errMsg)
				return -1, status.Error(codes.Internal, errMsg)
			}

			i++
		}
	}
}

func (s *ApiService) sendSubscribeTx(server apiPb.RpcNode_SubscribeServer,
	txs []*commonPb.Transaction, contractName string, txIds []string,
	txIdsMap map[string]struct{}, reqSender protocol.Role, reqSenderOrgId string) error {

	var (
		err error
	)

	for _, tx := range txs {
		if contractName == "" && len(txIds) == 0 {
			if err = s.doSendSubscribeTx(server, tx, reqSender, reqSenderOrgId); err != nil {
				return err
			}
			continue
		}

		if s.checkIsContinue(tx, contractName, txIds, txIdsMap) {
			continue
		}

		if err = s.doSendSubscribeTx(server, tx, reqSender, reqSenderOrgId); err != nil {
			return err
		}
	}

	return nil
}

func (s *ApiService) doSendSubscribeTx(server apiPb.RpcNode_SubscribeServer, tx *commonPb.Transaction,
	reqSender protocol.Role, reqSenderOrgId string) error {

	var (
		err    error
		errMsg string
		result *commonPb.SubscribeResult
	)

	isReqSenderLightNode := reqSender == protocol.RoleLight
	isTxRelatedToSender := (tx.Sender != nil) && reqSenderOrgId == tx.Sender.Signer.OrgId

	if result, err = s.getTxSubscribeResult(tx); err != nil {
		errMsg = fmt.Sprintf("get tx subscribe result failed, %s", err)
		s.log.Error(errMsg)
		return errors.New(errMsg)
	}

	if isReqSenderLightNode {
		if isTxRelatedToSender {
			if err := server.Send(result); err != nil {
				errMsg = fmt.Sprintf("send subscribe tx result failed, %s", err)
				s.log.Error(errMsg)
				return errors.New(errMsg)
			}
		}
	} else {
		if err := server.Send(result); err != nil {
			errMsg = fmt.Sprintf("send subscribe tx result failed, %s", err)
			s.log.Error(errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
}

func (s *ApiService) getTxSubscribeResult(tx *commonPb.Transaction) (*commonPb.SubscribeResult, error) {
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		errMsg := fmt.Sprintf("marshal tx info failed, %s", err)
		s.log.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	result := &commonPb.SubscribeResult{
		Data: txBytes,
	}

	return result, nil
}

func (s *ApiService) checkIsContinue(tx *commonPb.Transaction, contractName string, txIds []string,
	txIdsMap map[string]struct{}) bool {

	if contractName != "" && tx.Payload.ContractName != contractName {
		return true
	}

	if len(txIds) > 0 {
		_, ok := txIdsMap[tx.Payload.TxId]
		if !ok {
			return true
		}

		delete(txIdsMap, tx.Payload.TxId)
	}

	return false
}
