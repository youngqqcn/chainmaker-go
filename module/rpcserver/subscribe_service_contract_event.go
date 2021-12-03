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
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	protocol "chainmaker.org/chainmaker/protocol/v2"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//dealContractEventSubscription - deal contract event subscribe request
func (s *ApiService) dealContractEventSubscription(tx *commonPb.Transaction,
	server apiPb.RpcNode_SubscribeServer) error {

	var (
		err          error
		errMsg       string
		errCode      commonErr.ErrCode
		db           protocol.BlockchainStore
		payload      = tx.Payload
		startBlock   int64
		endBlock     int64
		contractName string
		topic        string
	)

	for _, kv := range payload.Parameters {
		if kv.Key == syscontract.SubscribeContractEvent_START_BLOCK.String() {
			startBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeContractEvent_END_BLOCK.String() {
			endBlock, err = bytehelper.BytesToInt64(kv.Value)
		} else if kv.Key == syscontract.SubscribeContractEvent_CONTRACT_NAME.String() {
			contractName = string(kv.Value)
		} else if kv.Key == syscontract.SubscribeContractEvent_TOPIC.String() {
			if kv.Value != nil {
				topic = string(kv.Value)
			}
		}

		if err != nil {
			errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_CONTRACT_EVENT
			errMsg = s.getErrMsg(errCode, err)
			s.log.Error(errMsg)
			return status.Error(codes.InvalidArgument, errMsg)
		}
	}

	if err = s.checkSubscribeContractEventPayload(startBlock, endBlock, contractName); err != nil {
		errCode = commonErr.ERR_CODE_CHECK_PAYLOAD_PARAM_SUBSCRIBE_CONTRACT_EVENT
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	s.log.Infof("Recv contract event subscribe request: [start:%d]/[end:%d]/[contractName:%s]/[topic:%s]",
		startBlock, endBlock, contractName, topic)

	chainId := tx.Payload.ChainId
	if db, err = s.chainMakerServer.GetStore(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_STORE
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	return s.doSendContractEvent(tx, db, server, startBlock, endBlock, contractName, topic)
}

func (s *ApiService) checkSubscribeContractEventPayload(startBlockHeight, endBlockHeight int64,
	contractName string) error {

	if startBlockHeight < -1 || endBlockHeight < -1 ||
		(endBlockHeight != -1 && startBlockHeight > endBlockHeight) {

		return errors.New("invalid start block height or end block height")
	}

	if contractName == "" {
		return errors.New("contractName can't be empty")
	}

	return nil
}

func (s *ApiService) doSendContractEvent(tx *commonPb.Transaction, db protocol.BlockchainStore,
	server apiPb.RpcNode_SubscribeServer, startBlock, endBlock int64,
	contractName string, topic string) error {

	var (
		alreadySendHistoryBlockHeight int64
		err                           error
	)

	if startBlock == -1 && endBlock == 0 {
		return status.Error(codes.OK, "OK")
	}

	// just send realtime contract event
	// == 0 for compatibility
	if (startBlock == -1 && endBlock == -1) || (startBlock == 0 && endBlock == 0) {
		return s.sendNewContractEvent(db, tx, server, startBlock, endBlock, contractName, topic, -1)
	}

	if startBlock != -1 {
		if alreadySendHistoryBlockHeight, err = s.doSendHistoryContractEvent(db, server, startBlock, endBlock,
			contractName, topic); err != nil {
			return err
		}
	}

	if startBlock == -1 {
		alreadySendHistoryBlockHeight = -1
	}

	if alreadySendHistoryBlockHeight == 0 {
		return status.Error(codes.OK, "OK")
	}

	return s.sendNewContractEvent(db, tx, server, startBlock, endBlock, contractName, topic,
		alreadySendHistoryBlockHeight)
}

func (s *ApiService) doSendHistoryContractEvent(db protocol.BlockchainStore, server apiPb.RpcNode_SubscribeServer,
	startBlock, endBlock int64, contractName, topic string) (int64, error) {

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

	// only send history contract event
	if endBlock > 0 && endBlock <= lastBlockHeight {
		_, err = s.sendHistoryContractEvent(db, server, startBlockHeight, endBlock, contractName, topic)

		if err != nil {
			s.log.Errorf("sendHistoryContractEvent failed, %s", err)
			return -1, err
		}

		return 0, status.Error(codes.OK, "OK")
	}

	alreadySendHistoryBlockHeight, err := s.sendHistoryContractEvent(db, server, startBlockHeight, endBlock,
		contractName, topic)

	if err != nil {
		s.log.Errorf("sendHistoryContractEvent failed, %s", err)
		return -1, err
	}

	s.log.Debugf("after sendHistoryContractEvent, alreadySendHistoryBlockHeight is %d",
		alreadySendHistoryBlockHeight)

	return alreadySendHistoryBlockHeight, nil
}

// sendHistoryContractEvent - send history contract event to subscriber
func (s *ApiService) sendHistoryContractEvent(store protocol.BlockchainStore,
	server apiPb.RpcNode_SubscribeServer,
	startBlockHeight, endBlockHeight int64,
	contractName, topic string) (int64, error) {

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

			if endBlockHeight > 0 && i > endBlockHeight {
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

			if err := s.sendSubscribeContractEvent(server, block, contractName, topic); err != nil {
				errMsg = fmt.Sprintf("send subscribe tx failed, %s", err)
				s.log.Error(errMsg)
				return -1, status.Error(codes.Internal, errMsg)
			}

			i++
		}
	}
}

func (s *ApiService) sendSubscribeContractEvent(server apiPb.RpcNode_SubscribeServer,
	block *commonPb.Block, contractName, topic string) error {

	var (
		err error
	)

	for _, tx := range block.Txs {
		var eventInfos commonPb.ContractEventInfoList
		for idx, event := range tx.Result.ContractResult.ContractEvent {
			if topic == "" || topic == event.Topic {
				if contractName != event.ContractName {
					continue
				}

				eventInfo := commonPb.ContractEventInfo{
					BlockHeight:     block.Header.BlockHeight,
					ChainId:         block.Header.ChainId,
					Topic:           event.Topic,
					TxId:            tx.Payload.TxId,
					EventIndex:      uint32(idx),
					ContractName:    event.ContractName,
					ContractVersion: event.ContractVersion,
					EventData:       event.EventData,
				}

				if eventInfo.BlockHeight != 0 {
					eventInfos.ContractEvents = append(eventInfos.ContractEvents, &eventInfo)
				}
			}
		}

		if err = s.doSendSubscribeContractEvent(server, eventInfos.ContractEvents, contractName, topic); err != nil {
			return err
		}
	}

	return nil
}

func (s *ApiService) doSendSubscribeContractEvent(server apiPb.RpcNode_SubscribeServer,
	contractEvents []*commonPb.ContractEventInfo, contractName, topic string) error {

	var (
		err    error
		errMsg string
		result *commonPb.SubscribeResult
	)

	sendContractEvents := []*commonPb.ContractEventInfo{}
	for _, EventInfo := range contractEvents {
		if EventInfo.ContractName != contractName || (topic != "" && EventInfo.Topic != topic) {
			continue
		}
		sendContractEvents = append(sendContractEvents, EventInfo)
	}

	if len(sendContractEvents) == 0 {
		return nil
	}

	if result, err = s.getContractEventSubscribeResult(sendContractEvents); err != nil {
		errMsg = fmt.Sprintf("get contract event subscribe result failed, %s", err)
		s.log.Error(errMsg)
		return errors.New(errMsg)
	}

	if err := server.Send(result); err != nil {
		errMsg = fmt.Sprintf("send subscribe contract event result failed, %s", err)
		s.log.Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (s *ApiService) sendNewContractEvent(store protocol.BlockchainStore, tx *commonPb.Transaction,
	server apiPb.RpcNode_SubscribeServer, startBlock, endBlock int64,
	contractName string, topic string, alreadySendHistoryBlockHeight int64) error {

	var (
		errCode            commonErr.ErrCode
		err                error
		errMsg             string
		eventSubscriber    *subscriber.EventSubscriber
		historyBlockHeight int64
	)

	eventCh := make(chan model.NewContractEvent)

	chainId := tx.Payload.ChainId
	if eventSubscriber, err = s.chainMakerServer.GetEventSubscribe(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_SUBSCRIBER
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	sub := eventSubscriber.SubscribeContractEvent(eventCh)
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-eventCh:
			contractEventInfoList := ev.ContractEventInfoList.ContractEvents

			blockHeight := int64(contractEventInfoList[0].BlockHeight)
			if endBlock > 0 && blockHeight > endBlock {
				return status.Error(codes.OK, "OK")
			}

			if alreadySendHistoryBlockHeight != -1 &&
				blockHeight > alreadySendHistoryBlockHeight {
				historyBlockHeight, err = s.sendHistoryContractEvent(store, server, alreadySendHistoryBlockHeight+1,
					blockHeight, contractName, topic)
				if err != nil {
					s.log.Errorf("send history contract event failed, %s", err)
					return err
				}

				if endBlock > 0 && historyBlockHeight >= endBlock {
					return status.Error(codes.OK, "OK")
				}

				alreadySendHistoryBlockHeight = -1
				continue
			}

			if err = s.doSendSubscribeContractEvent(server, contractEventInfoList, contractName, topic); err != nil {
				return status.Error(codes.Internal, errMsg)
			}

			if endBlock > 0 && blockHeight >= endBlock {
				return status.Error(codes.OK, "OK")
			}

		case <-server.Context().Done():
			return nil
		case <-s.ctx.Done():
			return nil
		}
	}
}

//func (s *ApiService) getContractEventSubscribeResult(contractEventsInfoList *commonPb.ContractEventInfoList) (
func (s *ApiService) getContractEventSubscribeResult(contractEvents []*commonPb.ContractEventInfo) (
	*commonPb.SubscribeResult, error) {

	eventBytes, err := proto.Marshal(&commonPb.ContractEventInfoList{
		ContractEvents: contractEvents,
	})

	if err != nil {
		errMsg := fmt.Sprintf("marshal contract event info failed, %s", err)
		s.log.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	result := &commonPb.SubscribeResult{
		Data: eventBytes,
	}

	return result, nil
}
