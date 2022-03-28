/*
   Created by guoxin in 2022/3/2 7:52 PM
*/
package tx_duplicate_function

import (
	"chainmaker.org/chainmaker-go/module/accesscontrol"
	native "chainmaker.org/chainmaker-go/test/chainconfig_test"
	"chainmaker.org/chainmaker/common/v2/ca"
	"chainmaker.org/chainmaker/common/v2/container"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"chainmaker.org/chainmaker/localconf/v2"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/test"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"chainmaker.org/chainmaker/store/v2"
	"chainmaker.org/chainmaker/store/v2/conf"
	"chainmaker.org/chainmaker/utils/v2"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	storeIns protocol.BlockchainStore

	RWPath = "test%v.txt"
)

const (
	ChainName = "chain1"

	// errors
	signFailedStr = "sign failed, %s"
	deadLineErr   = "WARN: client.call err: deadline"
)

func CheckFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
func DeleteFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		return
	}
}
func Delete(index int) {
	DeleteFile(pathPattern(index))
}
func Exist(index int) bool {
	return CheckFileExist(pathPattern(index))
}
func pathPattern(index int) string {
	return fmt.Sprintf(RWPath, index)
}

func ReadSlices(index int) (s []string) {
	f, err := ioutil.ReadFile(pathPattern(index))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(f, &s)
	if err != nil {
		panic(err)
	}
	return

}

func WriteSlices(s []string, index int) {
	marshal, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(pathPattern(index)) //创建文件
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(marshal) //写入文件(字节数组)
	if err != nil {
		panic(err)
	}
	fmt.Println("write success")

}

// SendTx 调用合约发送交易
func SendTx(txId string, userCrtPath, userKeyPath, IP string, Port int, caPath []string) {
	sk, member := getUserSK(1, userKeyPath, userCrtPath)
	resp, err := sendRequest(sk, member, true, &native.InvokeContractMsg{
		TxId:         txId,
		ChainId:      ChainName,
		TxType:       commonPb.TxType_INVOKE_CONTRACT,
		ContractName: "T",
		MethodName:   "P",
	}, userCrtPath, userKeyPath, strconv.Itoa(Port), IP, caPath)
	if err == nil {
		log.Printf("end tx resp: code:%d, msg:%s, payload:%+v\n", resp.Code, resp.Message, resp.ContractResult)
		log.Println()
	}
	if statusErr, ok := status.FromError(err); ok && statusErr.Code() == codes.DeadlineExceeded {
		log.Println(deadLineErr)
		log.Println()
	}
	time.Sleep(5 * time.Second)
}

func initStore() {
	var storeFactory store.Factory // nolint: typecheck
	err := container.Register(func() protocol.Logger { return &test.GoLogger{} }, container.Name("store"))
	if err != nil {
		log.Fatalf("Store Register log %v", err)
	}

	err = container.Register(localconf.ChainMakerConfig.GetP11Handle)
	if err != nil {
		log.Fatalf("Store Register GetP11Handle %v", err)
	}
	err = container.Register(storeFactory.NewStore,
		container.Parameters(map[int]interface{}{0: ChainName, 1: &conf.StorageConfig{
			StorePath: "./data/org1/ledgerData1",
			BlockDbConfig: &conf.DbConfig{
				Provider: "leveldb",
				LevelDbConfig: map[string]interface{}{
					"store_path": "./data/org1/blocks",
				},
			},
			StateDbConfig: &conf.DbConfig{
				Provider: "leveldb",
				LevelDbConfig: map[string]interface{}{
					"store_path": "./data/org1/statedb",
				},
			},
			HistoryDbConfig: &conf.HistoryDbConfig{
				DbConfig: conf.DbConfig{
					Provider: "leveldb",
					LevelDbConfig: map[string]interface{}{
						"store_path": "./data/org1/history",
					},
				},
				DisableKeyHistory:      false,
				DisableContractHistory: false,
				DisableAccountHistory:  false,
			},
			ResultDbConfig: &conf.DbConfig{
				Provider: "leveldb",
				LevelDbConfig: map[string]interface{}{
					"store_path": "./data/org1/result",
				},
			},
			DisableContractEventDB: true,
			Encryptor:              "sm4",
			EncryptKey:             "1234567890123456",
		}}),
		container.DependsOn(map[int]string{2: "store"}),
		container.Name(ChainName))
	if err != nil {
		log.Fatalf("Store Register %v", err)
	}
	err = container.Resolve(&storeIns, container.ResolveName(ChainName))
	if err != nil {
		log.Fatalf("Store Resolve %v", err)
	}
}

// 获取用户私钥
func getUserSK(orgIDNum int, keyPath, certPath string) (crypto.PrivateKey, *acPb.Member) {
	file, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	sk3, err := asym.PrivateKeyFromPEM(file, nil)
	if err != nil {
		panic(err)
	}
	file2, err := ioutil.ReadFile(certPath)
	if err != nil {
		panic(err)
	}
	sender := &acPb.Member{
		OrgId:      fmt.Sprintf("wx-org%d.chainmaker.org", orgIDNum),
		MemberInfo: file2,
		//IsFullCert: true,
	}
	return sk3, sender
}

func sendRequest(sk3 crypto.PrivateKey, sender *acPb.Member, isTls bool, msg *native.InvokeContractMsg, userCrtPath, userKeyPath, Port, IP string, caPath []string) (*commonPb.TxResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()

	conn, err := initGRPCConn(isTls, userCrtPath, userKeyPath, Port, IP, caPath)
	if err != nil {
		panic(err)
	}
	client := apiPb.NewRpcNodeClient(conn)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5*time.Second)))
	defer cancel()
	if msg.TxId == "" {
		msg.TxId = strconv.FormatInt(time.Now().UnixNano()/1e6, 10) + sdkutils.SeparatorString + uuid.GetUUID()
	}
	// 构造Header
	header := &commonPb.Payload{
		ChainId: msg.ChainId,
		//Sender:         sender,
		TxType:         msg.TxType,
		TxId:           msg.TxId,
		Timestamp:      time.Now().Unix(),
		ExpirationTime: 0,

		ContractName: msg.ContractName,
		Method:       msg.MethodName,
		Parameters:   msg.Pairs,
	}
	//payloadBytes, err := proto.Marshal(payload)
	//if err != nil {
	//	log.Fatalf(marshalFailedStr, err.Error())
	//}
	req := &commonPb.TxRequest{
		Payload: header,
		Sender:  &commonPb.EndorsementEntry{Signer: sender},
	}
	// 拼接后，计算Hash，对hash计算签名
	rawTxBytes, err := utils.CalcUnsignedTxRequestBytes(req)
	if err != nil {
		log.Fatalf("CalcUnsignedTxRequest failed in sendRequest, %s", err.Error())
	}
	signer := getSigner(sk3, sender)
	signBytes, err := signer.Sign(crypto.CRYPTO_ALGO_SHA256, rawTxBytes)
	if err != nil {
		log.Fatalf(signFailedStr, err.Error())
	}
	err = signer.Verify(crypto.CRYPTO_ALGO_SHA256, rawTxBytes, signBytes)
	if err != nil {
		panic(err)
	}
	req.Sender.Signature = signBytes
	fmt.Println(req.Payload.TxId)
	return client.SendRequest(ctx, req)
}

func initGRPCConn(useTLS bool, userCrtPath, userKeyPath, Port, IP string, caPath []string) (*grpc.ClientConn, error) {
	url := fmt.Sprintf("%s:%v", IP, Port)
	if useTLS {
		tlsClient := ca.CAClient{
			ServerName: "chainmaker.org",
			CaPaths:    caPath,
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

func getSigner(sk3 crypto.PrivateKey, sender *acPb.Member) protocol.SigningMember {
	skPEM, err := sk3.String()
	if err != nil {
		log.Fatalf("get sk PEM failed, %s", err.Error())
	}
	signer, err := accesscontrol.NewCertSigningMember("", sender, skPEM, "")
	if err != nil {
		panic(err)
	}
	return signer
}
