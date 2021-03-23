/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rpcserver

import (
	"chainmaker.org/chainmaker-go/blockchain"
	commonErr "chainmaker.org/chainmaker-go/common/errors"
	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker-go/monitor"
	apiPb "chainmaker.org/chainmaker-go/pb/protogo/api"
	commonPb "chainmaker.org/chainmaker-go/pb/protogo/common"
	configPb "chainmaker.org/chainmaker-go/pb/protogo/config"
	netPb "chainmaker.org/chainmaker-go/pb/protogo/net"
	"chainmaker.org/chainmaker-go/protocol"
	"chainmaker.org/chainmaker-go/utils"
	"chainmaker.org/chainmaker-go/vm/native"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

const (
	SYSTEM_CHAIN = "system_chain"
)

var _ apiPb.RpcNodeServer = (*ApiService)(nil)

// ApiService struct define
type ApiService struct {
	chainMakerServer      *blockchain.ChainMakerServer
	log                   *logger.CMLogger
	subscriberRateLimiter *rate.Limiter
	metricQueryCounter    *prometheus.CounterVec
	metricInvokeCounter   *prometheus.CounterVec

	ctx context.Context
}

// NewApiService - new ApiService object
func NewApiService(chainMakerServer *blockchain.ChainMakerServer, ctx context.Context) *ApiService {
	log := logger.GetLogger(logger.MODULE_RPC)

	tokenBucketSize := localconf.ChainMakerConfig.RpcConfig.SubscriberConfig.RateLimitConfig.TokenBucketSize
	tokenPerSecond := localconf.ChainMakerConfig.RpcConfig.SubscriberConfig.RateLimitConfig.TokenPerSecond

	var subscriberRateLimiter *rate.Limiter
	if tokenBucketSize >= 0 && tokenPerSecond >= 0 {
		if tokenBucketSize == 0 {
			tokenBucketSize = subscriberRateLimitDefaultTokenBucketSize
		}

		if tokenPerSecond == 0 {
			tokenPerSecond = subscriberRateLimitDefaultTokenPerSecond
		}

		subscriberRateLimiter = rate.NewLimiter(rate.Limit(tokenPerSecond), tokenBucketSize)
	}

	apiService := ApiService{
		chainMakerServer:      chainMakerServer,
		log:                   log,
		subscriberRateLimiter: subscriberRateLimiter,
		ctx:                   ctx,
	}

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		apiService.metricQueryCounter = monitor.NewCounterVec(monitor.SUBSYSTEM_RPCSERVER, "metric_query_request_counter",
			"query request counts metric", "chainId", "state")
		apiService.metricInvokeCounter = monitor.NewCounterVec(monitor.SUBSYSTEM_RPCSERVER, "metric_invoke_request_counter",
			"invoke request counts metric", "chainId", "state")
	}

	return &apiService
}

// SendRequest - deal received TxRequest
func (s *ApiService) SendRequest(ctx context.Context, req *commonPb.TxRequest) (*commonPb.TxResponse, error) {

	return s.invoke(&commonPb.Transaction{
		Header:           req.Header,
		RequestPayload:   req.Payload,
		RequestSignature: req.Signature,
		Result:           nil}, protocol.RPC), nil
}

// validate tx
func (s *ApiService) validate(tx *commonPb.Transaction) (errCode commonErr.ErrCode, errMsg string) {
	var (
		err error
		bc  *blockchain.Blockchain
	)

	_, err = s.chainMakerServer.GetChainConf(tx.Header.ChainId)
	if err != nil {
		errCode = commonErr.ERR_CODE_GET_CHAIN_CONF
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return
	}

	bc, err = s.chainMakerServer.GetBlockchain(tx.Header.ChainId)
	if err != nil {
		errCode = commonErr.ERR_CODE_GET_BLOCKCHAIN
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		return
	}

	if err = utils.VerifyTxWithoutPayload(tx, tx.Header.ChainId, bc.GetAccessControl()); err != nil {
		errCode = commonErr.ERR_CODE_TX_VERIFY_FAILED
		errMsg = fmt.Sprintf("%s, %s, txId:%s, sender:%s", errCode.String(), err.Error(), tx.Header.TxId,
			hex.EncodeToString(tx.Header.Sender.MemberInfo))
		s.log.Error(errMsg)
		return
	}

	return commonErr.ERR_CODE_OK, ""
}

func (s *ApiService) getErrMsg(errCode commonErr.ErrCode, err error) string {
	return fmt.Sprintf("%s, %s", errCode.String(), err.Error())
}

// invoke contract according to TxType
func (s *ApiService) invoke(tx *commonPb.Transaction, source protocol.TxSource) *commonPb.TxResponse {
	var (
		errCode commonErr.ErrCode
		errMsg  string
		resp    = &commonPb.TxResponse{}
	)

	if tx.Header.ChainId != SYSTEM_CHAIN {
		errCode, errMsg = s.validate(tx)
		if errCode != commonErr.ERR_CODE_OK {
			resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
			resp.Message = errMsg
			return resp
		}
	}

	switch tx.Header.TxType {
	case commonPb.TxType_QUERY_SYSTEM_CONTRACT, commonPb.TxType_QUERY_USER_CONTRACT:
		return s.dealQuery(tx, source)
	case commonPb.TxType_INVOKE_USER_CONTRACT, commonPb.TxType_UPDATE_CHAIN_CONFIG, commonPb.TxType_MANAGE_USER_CONTRACT, commonPb.TxType_INVOKE_SYSTEM_CONTRACT:
		return s.dealTransact(tx, source)
	default:
		return &commonPb.TxResponse{
			Code:    commonPb.TxStatusCode_INTERNAL_ERROR,
			Message: commonErr.ERR_CODE_TXTYPE.String(),
		}
	}
}

// dealQuery - deal query tx
func (s *ApiService) dealQuery(tx *commonPb.Transaction, source protocol.TxSource) *commonPb.TxResponse {
	var (
		err     error
		errMsg  string
		errCode commonErr.ErrCode
		payload commonPb.TransactPayload
		store   protocol.BlockchainStore
		vmMgr   protocol.VmManager
		resp    = &commonPb.TxResponse{}
	)

	chainId := tx.Header.ChainId

	if store, err = s.chainMakerServer.GetStore(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_STORE
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		resp.Message = errMsg
		return resp
	}

	if vmMgr, err = s.chainMakerServer.GetVmManager(chainId); err != nil {
		errCode = commonErr.ERR_CODE_GET_VM_MGR
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		resp.Message = errMsg
		return resp
	}

	if err = proto.Unmarshal(tx.RequestPayload, &payload); err != nil {
		errCode = commonErr.ERR_CODE_SYSTEM_CONTRACT_PB_UNMARSHAL
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		resp.Message = errMsg
		return resp
	}

	if chainId == SYSTEM_CHAIN {
		return s.dealSystemChainQuery(tx, vmMgr, source)
	}

	ctx := &txQuerySimContextImpl{
		tx:              tx,
		txReadKeyMap:    map[string]*commonPb.TxRead{},
		txWriteKeyMap:   map[string]*commonPb.TxWrite{},
		blockchainStore: store,
		vmManager:       vmMgr,
	}

	txResult, txStatusCode := vmMgr.RunContract(&commonPb.ContractId{ContractName: payload.ContractName}, payload.Method, nil, s.kvPair2Map(payload.Parameters), ctx, 0, tx.Header.TxType)
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		if txStatusCode == commonPb.TxStatusCode_SUCCESS && txResult.Code != commonPb.ContractResultCode_FAIL {
			s.metricQueryCounter.WithLabelValues(chainId, "true").Inc()
		} else {
			s.metricQueryCounter.WithLabelValues(chainId, "false").Inc()
		}
	}
	if txStatusCode != commonPb.TxStatusCode_SUCCESS {
		errCode = commonErr.ERR_CODE_INVOKE_CONTRACT
		errMsg = fmt.Sprintf("%d, %d, %s", txStatusCode, txResult.Code, txResult.Message)
		s.log.Error(errMsg)
		resp.Code = txStatusCode
		resp.Message = errMsg
		resp.ContractResult = txResult
		return resp
	}

	if txResult.Code == commonPb.ContractResultCode_FAIL {
		resp.Code = commonPb.TxStatusCode_CONTRACT_FAIL
		resp.Message = commonPb.TxStatusCode_CONTRACT_FAIL.String()
		resp.ContractResult = txResult
		return resp
	}

	resp.Code = commonPb.TxStatusCode_SUCCESS
	resp.Message = commonPb.TxStatusCode_SUCCESS.String()
	resp.ContractResult = txResult
	return resp
}

// dealSystemChainQuery - deal system chain query
func (s *ApiService) dealSystemChainQuery(tx *commonPb.Transaction, vmMgr protocol.VmManager, source protocol.TxSource) *commonPb.TxResponse {
	var (
		err     error
		errMsg  string
		errCode commonErr.ErrCode
		payload commonPb.TransactPayload
		resp    = &commonPb.TxResponse{}
	)

	chainId := tx.Header.ChainId

	if err = proto.Unmarshal(tx.RequestPayload, &payload); err != nil {
		errCode = commonErr.ERR_CODE_SYSTEM_CONTRACT_PB_UNMARSHAL
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		resp.Message = errMsg
		return resp
	}

	ctx := &txQuerySimContextImpl{
		tx:            tx,
		txReadKeyMap:  map[string]*commonPb.TxRead{},
		txWriteKeyMap: map[string]*commonPb.TxWrite{},
		vmManager:     vmMgr,
	}

	runtimeInstance := native.GetRuntimeInstance(chainId)
	txResult := runtimeInstance.Invoke(&commonPb.ContractId{
		ContractName: payload.ContractName,
	},
		payload.Method,
		nil,
		s.kvPair2Map(payload.Parameters),
		ctx,
	)

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		if txResult.Code != commonPb.ContractResultCode_FAIL {
			s.metricQueryCounter.WithLabelValues(chainId, "true").Inc()
		} else {
			s.metricQueryCounter.WithLabelValues(chainId, "false").Inc()
		}
	}

	if txResult.Code == commonPb.ContractResultCode_FAIL {
		resp.Code = commonPb.TxStatusCode_CONTRACT_FAIL
		resp.Message = commonPb.TxStatusCode_CONTRACT_FAIL.String()
		resp.ContractResult = txResult
		return resp
	}

	resp.Code = commonPb.TxStatusCode_SUCCESS
	resp.Message = commonPb.TxStatusCode_SUCCESS.String()
	resp.ContractResult = txResult
	return resp
}

// kvPair2Map - change []*commonPb.KeyValuePair to map[string]string
func (s *ApiService) kvPair2Map(kvPair []*commonPb.KeyValuePair) map[string]string {
	kvMap := make(map[string]string)

	for _, kv := range kvPair {
		kvMap[kv.Key] = kv.Value
	}

	return kvMap
}

// dealTransact - deal transact tx
func (s *ApiService) dealTransact(tx *commonPb.Transaction, source protocol.TxSource) *commonPb.TxResponse {
	var (
		err     error
		errMsg  string
		errCode commonErr.ErrCode
		resp    = &commonPb.TxResponse{}
	)

	// whether modify tx payload
	if localconf.ChainMakerConfig.DebugConfig.IsModifyTxPayload {
		tx.RequestPayload = append(tx.RequestPayload, byte(0)) // append zero byte
	}

	// spv logic
	if localconf.ChainMakerConfig.NodeConfig.Type == "spv" {
		return s.doSpvLogin(tx, resp)
	}

	err = s.chainMakerServer.AddTx(tx.Header.ChainId, tx, source)

	s.incInvokeCounter(tx.Header.ChainId, err)

	if err != nil {
		s.log.Warnf("Add tx failed, %s, chainId:%s, txId:%s",
			err.Error(), tx.Header.ChainId, tx.Header.TxId)

		errCode = commonErr.ERR_CODE_TX_ADD_FAILED
		errMsg = s.getErrMsg(errCode, err)
		s.log.Error(errMsg)
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		resp.Message = errMsg
		return resp

	}

	s.log.Debugf("Add tx success, chainId:%s, txId:%s", tx.Header.ChainId, tx.Header.TxId)

	errCode = commonErr.ERR_CODE_OK
	resp.Code = commonPb.TxStatusCode_SUCCESS
	resp.Message = errCode.String()

	return resp
}

func (s *ApiService) incInvokeCounter(chainId string, err error) {
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		if err == nil {
			s.metricInvokeCounter.WithLabelValues(chainId, "true").Inc()
		} else {
			s.metricInvokeCounter.WithLabelValues(chainId, "false").Inc()
		}
	}
}

func (s *ApiService) doSpvLogin(tx *commonPb.Transaction, resp *commonPb.TxResponse) *commonPb.TxResponse {
	var (
		err    error
		errMsg string
		store  protocol.BlockchainStore
	)

	chainId := tx.Header.ChainId

	if store, err = s.chainMakerServer.GetStore(chainId); err != nil {
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		errMsg = fmt.Sprintf("%s", err.Error())
		s.log.Error(errMsg)
		resp.Message = errMsg
		return resp
	}

	exist, err := store.TxExists(tx.Header.TxId)
	if err != nil {
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		errMsg = fmt.Sprintf("%s", err.Error())
		s.log.Error(errMsg)
		resp.Message = errMsg
		return resp
	}

	if exist {
		resp.Code = commonPb.TxStatusCode_SUCCESS
		resp.Message = commonErr.ERR_CODE_OK.String()
		return resp
	}

	txMsg, err := proto.Marshal(tx)
	if err != nil {
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		errMsg = fmt.Sprintf("%s", err.Error())
		s.log.Error(errMsg)
		resp.Message = errMsg
		return resp
	}

	netService, err := s.chainMakerServer.GetNetService(chainId)
	if err != nil {
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		errMsg = fmt.Sprintf("%s", err.Error())
		s.log.Error(errMsg)
		resp.Message = errMsg
		return resp
	}

	err = netService.BroadcastMsg(txMsg, netPb.NetMsg_TX)
	if err != nil {
		resp.Code = commonPb.TxStatusCode_INTERNAL_ERROR
		errMsg = fmt.Sprintf("%s", err.Error())
		s.log.Error(errMsg)
		resp.Message = errMsg
		return resp
	}

	resp.Code = commonPb.TxStatusCode_SUCCESS
	resp.Message = commonErr.ERR_CODE_OK.String()
	return resp
}

// RefreshLogLevelsConfig - refresh log level
func (s *ApiService) RefreshLogLevelsConfig(ctx context.Context, req *configPb.LogLevelsRequest) (*configPb.LogLevelsResponse, error) {
	if err := localconf.RefreshLogLevelsConfig(); err != nil {
		return &configPb.LogLevelsResponse{
			Code:    int32(1),
			Message: err.Error(),
		}, nil
	}
	return &configPb.LogLevelsResponse{
		Code: int32(0),
	}, nil
}

// UpdateDebugConfig - update debug config for test
func (s *ApiService) UpdateDebugConfig(ctx context.Context, req *configPb.DebugConfigRequest) (*configPb.DebugConfigResponse, error) {
	if err := localconf.UpdateDebugConfig(req.Pairs); err != nil {
		return &configPb.DebugConfigResponse{
			Code:    int32(1),
			Message: err.Error(),
		}, nil
	}
	return &configPb.DebugConfigResponse{
		Code: int32(0),
	}, nil
}

// CheckNewBlockChainConfig check new block chain config.
func (s *ApiService) CheckNewBlockChainConfig(context.Context, *configPb.CheckNewBlockChainConfigRequest) (*configPb.CheckNewBlockChainConfigResponse, error) {
	if err := localconf.CheckNewCmBlockChainConfig(); err != nil {
		return &configPb.CheckNewBlockChainConfigResponse{
			Code:    int32(1),
			Message: err.Error(),
		}, nil
	}
	return &configPb.CheckNewBlockChainConfigResponse{
		Code: int32(0),
	}, nil
}

// GetChainMakerVersion get chainmaker version by rpc request
func (s *ApiService) GetChainMakerVersion(ctx context.Context, req *configPb.ChainMakerVersionRequest) (*configPb.ChainMakerVersionResponse, error) {
	return &configPb.ChainMakerVersionResponse{
		Code:    int32(0),
		Version: s.chainMakerServer.Version(),
	}, nil
}