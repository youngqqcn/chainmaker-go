/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	acPb "chainmaker.org/chainmaker-go/pb/protogo/accesscontrol"
	apiPb "chainmaker.org/chainmaker-go/pb/protogo/api"
	commonPb "chainmaker.org/chainmaker-go/pb/protogo/common"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"chainmaker.org/chainmaker-go/accesscontrol"

	"chainmaker.org/chainmaker-go/common/json"

	"chainmaker.org/chainmaker-go/common/ca"
	"chainmaker.org/chainmaker-go/common/crypto"
	"chainmaker.org/chainmaker-go/common/crypto/asym"

	"chainmaker.org/chainmaker-go/protocol"
	"chainmaker.org/chainmaker-go/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	ip          string
	port        int
	chainId     string
	orgId       string
	caPaths     []string
	userCrtPath string
	userKeyPath string
	useTLS      bool
	dataFile    string
	startBlock  int64
	endBlock    int64
	withRwSet   bool
	txType      int32
	txIds       string

	conn   *grpc.ClientConn
	client apiPb.RpcNodeClient
	sk3    crypto.PrivateKey
)

func main() {
	var err error
	mainCmd := &cobra.Command{
		Use: "subscribe",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			conn, err = initGRPCConnect(true)
			if err != nil {
				panic(err)
			}

			client = apiPb.NewRpcNodeClient(conn)

			file, err := ioutil.ReadFile(userKeyPath)
			if err != nil {
				panic(err)
			}

			sk3, err = asym.PrivateKeyFromPEM(file, nil)
			if err != nil {
				panic(err)
			}
		},
	}

	mainCmd.AddCommand(SubscribeBlockCMD())
	mainCmd.AddCommand(SubscribeTxCMD())

	mainFlags := mainCmd.PersistentFlags()
	mainFlags.StringVarP(&ip, "ip", "i", "localhost", "specify ip")
	mainFlags.IntVarP(&port, "port", "p", 17988, "specify port")
	mainFlags.StringVarP(&userKeyPath, "userkey", "k", "../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.tls.key", "specify user key path")
	mainFlags.StringVarP(&userCrtPath, "user-crt", "c", "../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.tls.crt", "specify user crt path")
	mainFlags.StringArrayVarP(&caPaths, "ca-path", "P", []string{"../config/crypto-config/wx-org1.chainmaker.org/ca"}, "specify ca path")
	mainFlags.BoolVarP(&useTLS, "use-tls", "t", false, "specify whether use tls")
	mainFlags.StringVarP(&chainId, "chain-id", "C", "chain1", "specify chain id")
	mainFlags.StringVarP(&orgId, "org-id", "O", "wx-org1", "specify org id")
	mainFlags.StringVarP(&dataFile, "data-file", "f", "data.txt", "specify the data file to write blocks or tx")
	mainFlags.Int64VarP(&startBlock, "start-block", "s", 2, "specify the start block height to receive from, -1 means to receive until you stop the program")
	mainFlags.Int64VarP(&endBlock, "end-block", "e", -1, "specify the end block height to receive to, -1 means to receive until you stop the program")
	mainFlags.BoolVarP(&withRwSet, "withRWSet", "S", false, "specify withRWSet, true or false")
	mainFlags.Int32VarP(&txType, "tx-type", "T", -1, "specify transaction type you with to receive, -1 means all, other value from 0 to 7")
	mainFlags.StringVarP(&txIds, "tx-ids", "I", "", "specify the transaction ids, separated by comma, NOTICE: don't add space between ids")

	if mainCmd.Execute() != nil {
		return
	}
	if conn != nil {
		conn.Close()
	}
}

func initGRPCConnect(useTLS bool) (*grpc.ClientConn, error) {
	url := fmt.Sprintf("%s:%d", ip, port)

	if useTLS {
		tlsClient := ca.CAClient{
			ServerName: "chainmaker.org",
			CaPaths:    caPaths,
			CertFile:   userCrtPath,
			KeyFile:    userKeyPath,
		}

		c, err := tlsClient.GetCredentialsByCA()
		if err != nil {
			log.Fatalf("GetTLSCredentialsByCA err: %v", err)
			return nil, err
		}
		return grpc.Dial(url, grpc.WithTransportCredentials(*c))
	} else {
		return grpc.Dial(url, grpc.WithInsecure())
	}
}

func subscribeRequest(sk3 crypto.PrivateKey, client apiPb.RpcNodeClient, txType commonPb.TxType, _ string, payloadBytes []byte) (*commonPb.TxResponse, error) {

	req := generateReq(sk3, txType, payloadBytes)
	res, err := client.Subscribe(context.Background(), req)
	if err != nil {
		log.Fatalf("subscribe block failed, %s", err.Error())
	}

	f, err := os.OpenFile(dataFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("open data file failed, %s", err.Error())
	}
	defer f.Close()

	for {
		result, err := res.Recv()
		if err == io.EOF {
			log.Println("got eof and exit")
			break
		}
		if err != nil {
			log.Println(err)
			break
		}

		if txType == commonPb.TxType_SUBSCRIBE_BLOCK_INFO && recvBlock(f, result) != nil {
			break
		} else if recvTx(f, result) != nil {
			break
		}
	}

	return nil, err
}

func recvBlock(file *os.File, result *commonPb.SubscribeResult) error {
	var blockInfo commonPb.BlockInfo
	if err := proto.Unmarshal(result.Data, &blockInfo); err != nil {
		log.Println(err)
		return err
	}
	bytes, err := json.Marshal(blockInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	_, _ = file.Write(bytes)
	_, _ = file.WriteString("\n")
	blockHash := make([]byte, len(blockInfo.Block.Header.BlockHash)*2)
	hex.Encode(blockHash, blockInfo.Block.Header.BlockHash)
	fmt.Printf("Received a block at height:%d, chainId:%s, blockHash:%s\n",
		blockInfo.Block.Header.BlockHeight, chainId, blockHash)
	return nil
}
func recvTx(file *os.File, result *commonPb.SubscribeResult) error {
	var tx commonPb.Transaction
	if err := proto.Unmarshal(result.Data, &tx); err != nil {
		log.Println(err)
		return err
	}

	bytes, err := json.Marshal(tx)
	if err != nil {
		log.Println(err)
		return err
	}
	_, _ = file.Write(bytes)
	_, _ = file.WriteString("\n")

	fmt.Printf("Received a transaction, chainId:%s, txId:%s\n",
		tx.Header.ChainId, tx.Header.TxId)
	return nil
}

func generateReq(sk3 crypto.PrivateKey, txType commonPb.TxType, payloadBytes []byte) *commonPb.TxRequest {
	txId := utils.GetRandTxId()
	file, err := ioutil.ReadFile(userCrtPath)
	if err != nil {
		panic(err)
	}

	// 构造Sender
	sender := &acPb.SerializedMember{
		OrgId:      orgId,
		MemberInfo: file,
		IsFullCert: true,
	}

	// 构造Header
	header := &commonPb.TxHeader{
		ChainId:        chainId,
		Sender:         sender,
		TxType:         txType,
		TxId:           txId,
		Timestamp:      time.Now().Unix(),
		ExpirationTime: 0,
	}

	req := &commonPb.TxRequest{
		Header:    header,
		Payload:   payloadBytes,
		Signature: nil,
	}

	// 拼接后，计算Hash，对hash计算签名
	rawTxBytes, err := utils.CalcUnsignedTxRequestBytes(req)
	if err != nil {
		log.Fatalf("CalcUnsignedTxRequest failed, %s", err.Error())
	}

	signer := getSigner(sk3, sender)
	signBytes, err := signer.Sign("SM3", rawTxBytes)
	if err != nil {
		log.Fatalf("sign failed, %s", err.Error())
	}

	req.Signature = signBytes
	return req
}

func getSigner(sk3 crypto.PrivateKey, sender *acPb.SerializedMember) protocol.SigningMember {
	skPEM, err := sk3.String()
	if err != nil {
		log.Fatalf("get sk PEM failed, %s", err.Error())
	}

	m, err := accesscontrol.MockAccessControl().NewMemberFromCertPem(sender.OrgId, string(sender.MemberInfo))
	if err != nil {
		panic(err)
	}

	signer, err := accesscontrol.MockAccessControl().NewSigningMember(m, skPEM, "")
	if err != nil {
		panic(err)
	}
	return signer
}