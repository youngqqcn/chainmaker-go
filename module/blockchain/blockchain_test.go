/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package blockchain

import (
	"errors"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"

	"chainmaker.org/chainmaker-go/module/subscriber"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	msgbusMock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
)

const (
	id     = "QmQZn3pZCcuEf34FSvucqkvVJEvfzpNjQTk17HS6CYMR35"
	org1Id = "wx-org1"
)

func TestNewBlockchain(t *testing.T) {
	t.Log("TestNewBlockchain")

	var (
		genesis = "test"
		chainId = "123456"
		msgBus  msgbus.MessageBus
		net     protocol.Net
	)
	blockChain := NewBlockchain(genesis, chainId, msgBus, net)
	t.Log(blockChain)
}

func TestGetConsensusType(t *testing.T) {
	t.Log("TestGetConsensusType")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		res := blockchain.getConsensusType()
		t.Log(res)
	}

}

func TestGetAccessControl(t *testing.T) {
	t.Log("TestGetAccessControl")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		res := blockchain.GetAccessControl()
		t.Log(res)
	}
}

var consensusType = consensuspb.ConsensusType_SOLO

var chainConfig = &configpb.ChainConfig{
	Consensus: &configpb.ConsensusConfig{
		Type: consensusType,
		Nodes: []*configpb.OrgConfig{
			{
				OrgId:  org1Id,
				NodeId: []string{id},
			},
		},
	},
	Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
	Contract: &configpb.ContractConfig{
		EnableSqlSupport: true,
	},
	Block: &configpb.BlockConfig{
		BlockInterval: 5,
	},
}

func createBlockChain(t *testing.T) []*Blockchain {

	blockChainList := make([]*Blockchain, 0)

	for i := 0; i < 6; i++ {

		var (
			genesis         = "test" + strconv.Itoa(i)
			chainId         = "Chain" + strconv.Itoa(i)
			ctrl            = gomock.NewController(t)
			chainConf       = mock.NewMockChainConf(ctrl)
			ac              = mock.NewMockAccessControlProvider(ctrl)
			watcher         = mock.NewMockWatcher(ctrl)
			msgbusNew       = msgbusMock.NewMockMessageBus(ctrl)
			syncService     = mock.NewMockSyncService(ctrl)
			coreEngine      = mock.NewMockCoreEngine(ctrl)
			blockChainStore = mock.NewMockBlockchainStore(ctrl)
			//blockVerifier = mock.NewMockBlockVerifier(ctrl)
			//blockCommitter = mock.NewMockBlockCommitter(ctrl)
			txpool          = mock.NewMockTxPool(ctrl)
			snapshotManager = mock.NewMockSnapshotManager(ctrl)
			singMemer       = mock.NewMockSigningMember(ctrl)
			ledgerCache     = mock.NewMockLedgerCache(ctrl)
			vmMgr           = mock.NewMockVmManager(ctrl)
			proposalCache   = mock.NewMockProposalCache(ctrl)
			net             = mock.NewMockNet(ctrl)
			netService      = mock.NewMockNetService(ctrl)
			consensus       = mock.NewMockConsensusEngine(ctrl)
		)

		net.EXPECT().Start().AnyTimes().Return(nil)
		msgbusNew.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
		msgbusNew.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
		msgbusNew.EXPECT().PublishSafe(gomock.Any(), gomock.Any()).AnyTimes()

		blockchain := NewBlockchain(genesis, chainId, msgbusNew, net)

		consensusType = getConsensusType(i)

		chainConf.EXPECT().AddWatch(gomock.Any()).AnyTimes().Return()
		chainConf.EXPECT().AddVmWatch(gomock.Any()).AnyTimes().Return()
		chainConf.EXPECT().Init().AnyTimes().Return(nil)
		chainConf.EXPECT().GetChainConfigFromFuture(gomock.Any()).AnyTimes().Return(nil, nil)
		chainConf.EXPECT().GetChainConfigAt(gomock.Any()).AnyTimes().Return(nil, nil)
		chainConf.EXPECT().GetConsensusNodeIdList().AnyTimes().Return(nil, nil)
		chainConf.EXPECT().CompleteBlock(gomock.Any()).AnyTimes().Return(nil)

		ac.EXPECT().GetHashAlg().AnyTimes().Return(genesis) // 随机id
		watcher.EXPECT().Watch(chainConf).AnyTimes().Return(nil)
		coreEngine.EXPECT().GetBlockVerifier().AnyTimes().Return(nil)
		coreEngine.EXPECT().GetBlockCommitter().AnyTimes().Return(nil)
		coreEngine.EXPECT().Start().AnyTimes()
		syncService.EXPECT().Start().AnyTimes().Return(nil)

		blockChainStore.EXPECT().ReadObject(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().ArchiveBlock(gomock.Any()).AnyTimes().Return(nil)
		blockChainStore.EXPECT().Close().AnyTimes().Return(nil)
		blockChainStore.EXPECT().BeginDbTransaction(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().BlockExists(gomock.Any()).AnyTimes().Return(false, nil)
		blockChainStore.EXPECT().CommitDbTransaction(gomock.Any()).AnyTimes().Return(nil)
		blockChainStore.EXPECT().CreateDatabase(gomock.Any()).AnyTimes().Return(nil)
		blockChainStore.EXPECT().DropDatabase(gomock.Any()).AnyTimes().Return(nil)
		blockChainStore.EXPECT().ExecDdlSql(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		blockChainStore.EXPECT().GetAccountTxHistory(gomock.Any()).AnyTimes().Return(nil, nil)

		block := createNewBlock(uint64(i), int64(i), chainId)

		blockChainStore.EXPECT().GetBlock(gomock.Any()).AnyTimes().Return(block, nil)
		blockChainStore.EXPECT().GetArchivedPivot().AnyTimes().Return(uint64(22))
		blockChainStore.EXPECT().GetBlockByHash(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetBlockHeaderByHeight(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetBlockByTx(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetContractByName(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetContractBytecode(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetContractTxHistory(gomock.Any()).AnyTimes().Return(nil, nil)
		blockChainStore.EXPECT().GetDBHandle(gomock.Any()).AnyTimes().Return(nil)
		//blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, nil)

		if i < 3 {

			blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, nil)
			chainConfig.AuthType = protocol.PermissionedWithCert

			chainConfig.Snapshot = &configpb.SnapshotConfig{
				EnableEvidence: true,
			}

			consensus.EXPECT().Start().AnyTimes().Return(nil)
			consensus.EXPECT().Stop().AnyTimes().Return(nil)
			netService.EXPECT().Start().AnyTimes().Return(nil)
			netService.EXPECT().Stop().AnyTimes().Return(nil)
			singMemer.EXPECT().GetMember().AnyTimes().Return(nil, nil)

			txpool.EXPECT().Start().AnyTimes().Return(nil)
			blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, nil)

		} else if i < 4 {

			blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, errors.New("blockChainStore err test"))
			chainConfig.AuthType = protocol.Identity

			chainConfig.Snapshot = &configpb.SnapshotConfig{
				EnableEvidence: false,
			}
			consensus.EXPECT().Start().AnyTimes().Return(errors.New("consensus start test err msg"))
			consensus.EXPECT().Stop().AnyTimes().Return(nil)
			netService.EXPECT().Start().AnyTimes().Return(errors.New("netService start test err msg"))
			netService.EXPECT().Stop().AnyTimes().Return(nil)
			singMemer.EXPECT().GetMember().AnyTimes().Return(nil, errors.New("singMemer GetMember test err msg"))
			txpool.EXPECT().Start().AnyTimes().Return(errors.New("txpool start test err msg"))
			blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, errors.New("blockChainStore GetDbTransaction test err msg"))

		} else if i < 5 {

			blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(nil, errors.New("blockChainStore err test"))
			chainConfig.AuthType = protocol.Public

			chainConfig.Snapshot = &configpb.SnapshotConfig{
				EnableEvidence: true,
			}
			consensus.EXPECT().Start().AnyTimes().Return(nil)
			consensus.EXPECT().Stop().AnyTimes().Return(errors.New("consensus stop test err msg"))
			netService.EXPECT().Start().AnyTimes().Return(nil)
			netService.EXPECT().Stop().AnyTimes().Return(errors.New("netService start test err msg"))
			singMemer.EXPECT().GetMember().AnyTimes().Return(nil, nil)
			txpool.EXPECT().Start().AnyTimes().Return(nil)
			blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, nil)

		} else {

			blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(nil, nil)
			chainConfig.AuthType = protocol.PermissionedWithKey

			chainConfig.Snapshot = &configpb.SnapshotConfig{
				EnableEvidence: false,
			}
			consensus.EXPECT().Start().AnyTimes().Return(errors.New("consensus start test err msg"))
			consensus.EXPECT().Stop().AnyTimes().Return(errors.New("consensus stop test err msg"))
			netService.EXPECT().Start().AnyTimes().Return(errors.New("netService start test err msg"))
			netService.EXPECT().Stop().AnyTimes().Return(errors.New("netService start test err msg"))
			singMemer.EXPECT().GetMember().AnyTimes().Return(nil, nil)
			txpool.EXPECT().Start().AnyTimes().Return(errors.New("txpool start test err msg"))
			blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, errors.New("blockChainStore GetDbTransaction test err msg"))

		}

		txpool.EXPECT().Stop().AnyTimes()

		// 期望生成的 chainConf

		chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
		//chainConf.EXPECT().ChainConfig().AnyTimes().Return(getChainConfig()) // TODO

		coreEngine.EXPECT().Stop().AnyTimes()
		syncService.EXPECT().Stop().AnyTimes()

		err := chainConf.Init()

		if err != nil {
			t.Log(err)
		}

		blockchain.coreEngine = coreEngine
		blockchain.syncServer = syncService
		blockchain.netService = netService
		blockchain.identity = singMemer
		blockchain.ledgerCache = ledgerCache
		blockchain.vmMgr = vmMgr
		blockchain.proposalCache = proposalCache
		blockchain.eventSubscriber = &subscriber.EventSubscriber{}
		blockchain.store = blockChainStore
		blockchain.ac = ac // ac 赋值
		blockchain.txPool = txpool
		blockchain.chainConf = chainConf
		blockchain.snapshotManager = snapshotManager
		blockchain.consensus = consensus
		blockchain.coreEngine = coreEngine

		// 合并列表
		blockChainList = append(blockChainList, blockchain)
	}

	return blockChainList
}

// 不同情况的 consensusType
//consensusType := consensuspb.ConsensusType_SOLO // 0-SOLO,1-TBFT,2-MBFT,3-MAXBFT,4-RAFT,10-POW
func getConsensusType(i int) consensuspb.ConsensusType {

	switch i {
	case 0:
		consensusType = consensuspb.ConsensusType_SOLO
	case 1:
		consensusType = consensuspb.ConsensusType_SOLO // TBFT NOT FOUND TODO
	case 2:
		consensusType = consensuspb.ConsensusType_SOLO // MBFT NOT FOUND TODO
	case 3:
		consensusType = consensuspb.ConsensusType_MAXBFT
	case 4:
		consensusType = consensuspb.ConsensusType_RAFT
	case 5:
		consensusType = consensuspb.ConsensusType_SOLO // ConsensusType_POW NOT FOUND TODO
	default:
		consensusType = consensuspb.ConsensusType_SOLO
	}

	return consensusType
}

func createNewBlock(height uint64, timeStamp int64, chainId string) *commonPb.Block {
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			BlockHeight:    height,
			PreBlockHash:   nil,
			BlockHash:      nil,
			BlockVersion:   0,
			DagHash:        nil,
			RwSetRoot:      nil,
			BlockTimestamp: timeStamp,
			Proposer:       &acPb.Member{MemberInfo: []byte{1, 2, 3}},
			ConsensusArgs:  nil,
			TxCount:        0,
			Signature:      nil,
			ChainId:        chainId,
		},
		Dag: &commonPb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}
	block.Header.PreBlockHash = nil
	return block
}

/*
TODO getChainConfig
type InvokeContractMsg struct {
  TxType       commonPb.TxType
  ChainId      string
  TxId         string
  ContractName string
  MethodName   string
  Pairs        []*commonPb.KeyValuePair
}

func getChainConfig() *configpb.ChainConfig {
  conn, err := InitGRPCConnect(isTls)
  if err != nil {
    panic(err)
  }
  client := apiPb.NewRpcNodeClient(conn)

  fmt.Println("============ get chain config ============")
  // 构造Payload
  //pair := &commonPb.KeyValuePair{Key: "height", Value: strconv.FormatInt(1, 10)}
  var pairs []*commonPb.KeyValuePair
  //Pairs = append(Pairs, pair)

  sk, member := GetUserSK(1)
  resp, err := QueryRequest(sk, member, &client, &InvokeContractMsg{TxType: commonPb.TxType_QUERY_CONTRACT, ChainId: CHAIN1,
    ContractName: syscontract.SystemContract_CHAIN_CONFIG.String(), MethodName: syscontract.ChainConfigFunction_GET_CHAIN_CONFIG.String(), Pairs: pairs})
  if err == nil {
    if resp.Code != commonPb.TxStatusCode_SUCCESS {
      panic(resp.Message)
    }
    result := &configPb.ChainConfig{}
    err = proto.Unmarshal(resp.ContractResult.Result, result)
    data, _ := json.MarshalIndent(result, "", "\t")
    fmt.Printf("send tx resp: code:%d, msg:%s, chainConfig:%s\n", resp.Code, resp.Message, data)
    fmt.Printf("\n============ get chain config end============\n\n\n")
    return result
  }
  if statusErr, ok := status.FromError(err); ok && statusErr.Code() == codes.DeadlineExceeded {
    fmt.Println("WARN: client.call err: deadline")
    return nil
  }
  fmt.Printf("ERROR: client.call err: %v\n", err)
  return nil
}

// 获取用户私钥
func GetUserSK(index int) (crypto.PrivateKey, *acPb.Member) {
  numStr := strconv.Itoa(index)

  keyPath := fmt.Sprintf(UserSignKeyPathFmt, numStr)
  file, err := ioutil.ReadFile(keyPath)
  if err != nil {
    panic(err)
  }
  sk3, err := asym.PrivateKeyFromPEM(file, nil)
  if err != nil {
    panic(err)
  }
  certPath := fmt.Sprintf(UserSignCrtPathFmt, numStr)
  file2, err := ioutil.ReadFile(certPath)
  if err != nil {
    panic(err)
  }

  sender := &acPb.Member{
    OrgId:      fmt.Sprintf(OrgIdFormat, numStr),
    MemberInfo: file2,
    ////IsFullCert: true,
  }

  return sk3, sender
}

var (
  err    error
  CHAIN1 = "chain1"
  IP     = "localhost"
  Port   = 12301

  certPathPrefix = "../../config"
  //certPathPrefix     = "../../build"
  WasmPath           = "../wasm/counter-go.wasm"
  OrgIdFormat        = "wx-org%s.chainmaker.org"
  UserKeyPathFmt     = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/client1/client1.tls.key"
  UserCrtPathFmt     = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/client1/client1.tls.crt"
  UserSignKeyPathFmt = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/client1/client1.sign.key"
  UserSignCrtPathFmt = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/client1/client1.sign.crt"
  //UserSignKeyPathFmt  = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/light1/light1.sign.key"
  //UserSignCrtPathFmt  = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/light1/light1.sign.crt"
  AdminSignKeyPathFmt = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/admin1/admin1.sign.key"
  AdminSignCrtPathFmt = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/admin1/admin1.sign.crt"

  DefaultUserKeyPath = fmt.Sprintf(UserKeyPathFmt, "1")
  DefaultUserCrtPath = fmt.Sprintf(UserCrtPathFmt, "1")
  DefaultOrgId       = fmt.Sprintf(OrgIdFormat, "1")

  // caPaths    = []string{"D:/develop/workspace/chainMaker/chainmaker-go/build/crypto-config/wx-org5.chainmaker.org/ca"}
  caPaths    = []string{certPathPrefix + "/crypto-config/wx-org1.chainmaker.org/ca"}
  prePathFmt = certPathPrefix + "/crypto-config/wx-org%s.chainmaker.org/user/admin1/"

  isTls = true

  marshalErr    = "marshal payload failed, %s"
  signFailedErr = "sign failed, %s"
)

func QueryRequest(sk3 crypto.PrivateKey, sender *acPb.Member, client *apiPb.RpcNodeClient, msg *InvokeContractMsg) (*commonPb.TxResponse, error) {
  ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5*time.Second)))
  defer cancel()

  if msg.TxId == "" {
    msg.TxId = utils.GetRandTxId()
  }

  // 构造Header
  header := &commonPb.Payload{
    ChainId: msg.ChainId,
    //Sender:         sender,
    TxType:         msg.TxType,
    TxId:           msg.TxId,
    Timestamp:      time.Now().Unix(),
    ExpirationTime: 0,
    ContractName:   msg.ContractName,
    Method:         msg.MethodName,
    Parameters:     msg.Pairs,
  }

  req := &commonPb.TxRequest{
    Payload: header,
    Sender:  &commonPb.EndorsementEntry{Signer: sender},
  }

  // 拼接后，计算Hash，对hash计算签名
  rawTxBytes, err := utils.CalcUnsignedTxRequestBytes(req)
  if err != nil {
    log.Fatalf("CalcUnsignedTxRequest failed in QueryRequest, %s", err.Error())
  }

  signer := getSigner(sk3, sender)
  //signBytes, err := signer.Sign("SHA256", rawTxBytes)
  signBytes, err := signer.Sign(crypto.CRYPTO_ALGO_SHA256, rawTxBytes)
  if err != nil {
    log.Fatalf(signFailedErr, err.Error())
  }

  req.Sender.Signature = signBytes

  return (*client).SendRequest(ctx, req)
}

func InitGRPCConnect(useTLS bool) (*grpc.ClientConn, error) {
  url := fmt.Sprintf("%s:%d", IP, Port)

  if useTLS {
    tlsClient := ca.CAClient{
      ServerName: "chainmaker.org",
      CaPaths:    caPaths,
      CertFile:   DefaultUserCrtPath,
      KeyFile:    DefaultUserKeyPath,
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
*/
