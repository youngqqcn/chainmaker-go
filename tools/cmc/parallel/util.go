/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package parallel

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"chainmaker.org/chainmaker/common/v2/crypto"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"chainmaker.org/chainmaker/utils/v2"
)

const GRPCMaxCallRecvMsgSize = 16 * 1024 * 1024

func constructQueryPayload(chainId, contractName, method string, pairs []*commonPb.KeyValuePair, gasLimit uint64) (*commonPb.Payload, error) {
	payload := &commonPb.Payload{
		ContractName: contractName,
		Method:       method,
		Parameters:   pairs,
		TxId:         "", //Query不需要TxId
		TxType:       commonPb.TxType_QUERY_CONTRACT,
		ChainId:      chainId,
	}
	// gas limit
	if gasLimit > 0 {
		var limit = &commonPb.Limit{GasLimit: gasLimit}
		payload.Limit = limit
	}

	return payload, nil
}
func constructInvokePayload(chainId, contractName, method string, pairs []*commonPb.KeyValuePair, gasLimit uint64) (*commonPb.Payload, error) {
	payload := &commonPb.Payload{
		ContractName:   contractName,
		Method:         method,
		Parameters:     pairs,
		TxId:           utils.GetRandTxId(),
		TxType:         commonPb.TxType_INVOKE_CONTRACT,
		ChainId:        chainId,
		Timestamp:      time.Now().Unix(),
		ExpirationTime: 0,
	}
	// gas limit
	if gasLimit > 0 {
		var limit = &commonPb.Limit{GasLimit: gasLimit}
		payload.Limit = limit
	}

	return payload, nil
}

//func initGRPCConnect(useTLS bool) (*grpc.ClientConn, error) {
//	url := fmt.Sprintf("%s:%d", ip, port)
//
//	if useTLS {
//		tlsClient := ca.CAClient{
//			ServerName: "chainmaker.org",
//			CaPaths:    caPaths,
//			CertFile:   userCrtPath,
//			KeyFile:    userKeyPath,
//		}
//
//		c, err := tlsClient.GetCredentialsByCA()
//		if err != nil {
//			return nil, err
//		}
//		return grpc.Dial(url, grpc.WithTransportCredentials(*c),
//			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(GRPCMaxCallRecvMsgSize)))
//	} else {
//		return grpc.Dial(url, grpc.WithInsecure(),
//			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(GRPCMaxCallRecvMsgSize)))
//	}
//}

func acSign(msg *commonPb.Payload) ([]*commonPb.EndorsementEntry, error) {
	var adminKeys []string
	var adminCrts []string
	var adminOrgs []string
	if authType == sdk.Public {
		if adminSignKeys != "" {
			adminKeys = strings.Split(adminSignKeys, ",")
		}
		if len(adminKeys) == 0 {
			return nil, errors.New("admin keys is empty")
		}
	} else if authType == sdk.PermissionedWithKey {
		if adminSignKeys != "" {
			adminKeys = strings.Split(adminSignKeys, ",")
		}
		if orgIds != "" {
			adminOrgs = strings.Split(orgIds, ",")
		}
		if len(adminKeys) != len(adminOrgs) {
			return nil, errors.New("admin key len is not equal to orgId len")
		}
	} else {
		if adminSignKeys != "" {
			adminKeys = strings.Split(adminSignKeys, ",")
		}
		if adminSignCrts != "" {
			adminCrts = strings.Split(adminSignCrts, ",")
		}
		if len(adminKeys) != len(adminCrts) {
			return nil, errors.New("admin key len is not equal to crt len")
		}
	}

	hashType, err := getHashType(hashAlgo)
	if err != nil {
		return nil, err
	}
	endorsers := make([]*commonPb.EndorsementEntry, len(adminKeys))
	for i := range adminKeys {
		var e *commonPb.EndorsementEntry
		var err error
		if authType == sdk.PermissionedWithCert {
			e, err = sdkutils.MakeEndorserWithPath(adminKeys[i], adminCrts[i], msg)
		} else if authType == sdk.PermissionedWithKey {
			e, err = sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				hashType,
				adminOrgs[i],
				msg,
			)
		} else {
			e, err = sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				hashType,
				"",
				msg,
			)
		}
		if err != nil {
			return nil, err
		}
		endorsers[i] = e
	}
	return endorsers, nil
}

func getHashType(hashType string) (crypto.HashType, error) {
	if t, ok := crypto.HashAlgoMap[hashType]; ok {
		return t, nil
	}
	return 0, fmt.Errorf("unknown hash algo %s", hashType)
}
