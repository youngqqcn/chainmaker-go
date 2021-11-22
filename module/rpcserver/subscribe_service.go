/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcserver

import (
	"errors"
	"fmt"

	commonErr "chainmaker.org/chainmaker/common/v2/errors"
	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	protocol "chainmaker.org/chainmaker/protocol/v2"
	utils "chainmaker.org/chainmaker/utils/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// TRUE true string
	TRUE = "true"
)

// Subscribe - deal block/tx/contracEvent subscribe request
func (s *ApiService) Subscribe(req *commonPb.TxRequest, server apiPb.RpcNode_SubscribeServer) error {
	var (
		errCode commonErr.ErrCode
		errMsg  string
	)

	tx := &commonPb.Transaction{
		Payload:   req.Payload,
		Sender:    req.Sender,
		Endorsers: req.Endorsers,
		Result:    nil}

	errCode, errMsg = s.validate(tx)
	if errCode != commonErr.ERR_CODE_OK {
		return status.Error(codes.Unauthenticated, errMsg)
	}

	switch req.Payload.Method {
	case syscontract.SubscribeFunction_SUBSCRIBE_BLOCK.String():
		return s.dealBlockSubscription(tx, server)
	case syscontract.SubscribeFunction_SUBSCRIBE_TX.String():
		return s.dealTxSubscription(tx, server)
	case syscontract.SubscribeFunction_SUBSCRIBE_CONTRACT_EVENT.String():
		return s.dealContractEventSubscription(tx, server)
	}

	return nil
}

func (s *ApiService) checkAndGetLastBlockHeight(store protocol.BlockchainStore,
	payloadStartBlockHeight int64) (int64, error) {

	var (
		err             error
		errMsg          string
		errCode         commonErr.ErrCode
		lastBlock       *commonPb.Block
		lastBlockHeight uint64
	)

	if lastBlock, err = store.GetLastBlock(); err != nil {
		errCode = commonErr.ERR_CODE_GET_LAST_BLOCK
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return -1, status.Error(codes.Internal, errMsg)
	}

	lastBlockHeight = lastBlock.Header.BlockHeight

	if int64(lastBlockHeight) < payloadStartBlockHeight {
		errMsg = fmt.Sprintf("payload start block height:%d > last block height:%d",
			payloadStartBlockHeight, lastBlockHeight)

		s.log.Error(errMsg)
		return -1, status.Error(codes.InvalidArgument, errMsg)
	}

	return int64(lastBlock.Header.BlockHeight), nil
}

func (s *ApiService) getRateLimitToken() error {
	if s.subscriberRateLimiter != nil {
		if err := s.subscriberRateLimiter.Wait(s.ctx); err != nil {
			errMsg := fmt.Sprintf("subscriber rateLimiter wait token failed, %s", err.Error())
			s.log.Error(errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
}

// checkSubscribeBlockHeight - check subscriber payload info
func (s *ApiService) checkSubscribeBlockHeight(startBlockHeight, endBlockHeight int64) error {
	if startBlockHeight < -1 || endBlockHeight < -1 ||
		(endBlockHeight != -1 && startBlockHeight > endBlockHeight) {

		return errors.New("invalid start block height or end block height")
	}

	return nil
}

func (s *ApiService) getRoleFromTx(tx *commonPb.Transaction) (protocol.Role, error) {
	bc, err := s.chainMakerServer.GetBlockchain(tx.Payload.ChainId)
	if err != nil {
		errCode := commonErr.ERR_CODE_GET_BLOCKCHAIN
		errMsg := s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return "", err
	}

	ac := bc.GetAccessControl()
	return utils.GetRoleFromTx(tx, ac)
}
