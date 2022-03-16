/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
)

func TestInitNet(t *testing.T) {
	t.Log("TestInitNet")
	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.initNet()

		if err != nil {
			t.Log(err)
		}
	}

}

func TestInitBlockchains(t *testing.T) {
	t.Log("TestInitBlockchains")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.initBlockchains()

		if err != nil {
			t.Log(err)
		}
	}
}

func TestStart(t *testing.T) {
	t.Log("TestStart")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}
	}
}

func TestStop(t *testing.T) {
	t.Log("TestStop")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		server.Stop()
	}
}

func TestAddTx(t *testing.T) {
	t.Log("TestAddTx")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		server.AddTx("123456", &common.Transaction{
			//Payload:   "Payload",
			//Sender:    req.Sender,
			//Endorsers: req.Endorsers,
			Result: nil}, protocol.RPC)
	}
}

func TestGetStore(t *testing.T) {
	t.Log("TestGetStore")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		bkStore, err := server.GetStore("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(bkStore)
	}

}

func TestGetChainConf(t *testing.T) {
	t.Log("TestGetChainConf")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		chainConf, err := server.GetChainConf("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(chainConf)
	}

}

func TestGetAllChainConf(t *testing.T) {
	t.Log("TestGetAllChainConf")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		chainAllConf, err := server.GetAllChainConf()

		if err != nil {
			t.Log(err)
		}

		t.Log(chainAllConf)
	}

}

func TestGetVmManager(t *testing.T) {
	t.Log("TestGetVmManager")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		vmMgr, err := server.GetVmManager("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(vmMgr)
	}

}

func TestGetEventSubscribe(t *testing.T) {
	t.Log("TestGetEventSubscribe")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		sub, err := server.GetEventSubscribe("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(sub)
	}

}

func TestGetNetService(t *testing.T) {
	t.Log("TestGetNetService")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		netService, err := server.GetNetService("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(netService)
	}
}

func TestGetBlockchain(t *testing.T) {
	t.Log("TestGetBlockchain")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		block, err := server.GetBlockchain("123456")

		if err != nil {
			t.Log(err)
		}

		t.Log(block)
	}
}

func TestGetAllAC(t *testing.T) {
	t.Log("TestGetAllAC")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		acProvider, err := server.GetAllAC()

		if err != nil {
			t.Log(err)
		}

		t.Log(acProvider)
	}

}

func TestVersion(t *testing.T) {
	t.Log("TestVersion")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.Start()

		if err != nil {
			t.Log(err)
		}

		version := server.Version()

		t.Log(version)
	}

}

//

/*func TestInitBlockchain(t *testing.T)  {
	t.Log("TestInitBlockchain")

	chainmakerServer := ChainMakerServer{}
	err := chainmakerServer.initBlockchain("12344", "666")

	if err != nil {
		t.Log(err)
	}
}*/

// TODO
/*func TestNewBlockchainTaskListener(t *testing.T)  {
	t.Log("TestInitBlockchains")

	chainmakerServer := ChainMakerServer{}
	chainmakerServer.newBlockchainTaskListener()
	timer := time.NewTimer(3 * time.Second)
	<-timer.C
}*/

func makeChainServer(t *testing.T) []*ChainMakerServer {

	serverList := make([]*ChainMakerServer, 0)
	for i := 0; i < 4; i++ {
		var (
			ctrl = gomock.NewController(t)
			net  = mock.NewMockNet(ctrl)
			//blockChain = mock.NewMockSyncService(ctrl)
		)

		if i == 0 {
			net.EXPECT().Start().AnyTimes().Return(errors.New("server start test err msg"))
			net.EXPECT().Stop().AnyTimes().Return(nil)
		} else if i == 1 {
			net.EXPECT().Start().AnyTimes().Return(errors.New("server start test err msg"))
			net.EXPECT().Stop().AnyTimes().Return(errors.New("server stop test err msg"))
		} else if i == 2 {
			net.EXPECT().Start().AnyTimes().Return(nil)
			net.EXPECT().Stop().AnyTimes().Return(errors.New("server stop test err msg"))
		} else {
			net.EXPECT().Start().AnyTimes().Return(nil)
			net.EXPECT().Stop().AnyTimes().Return(nil)
		}

		//net.EXPECT().Start().AnyTimes().Return(nil)
		//net.EXPECT().Stop().AnyTimes().Return(nil)
		chainMaker := &ChainMakerServer{
			net:         net,
			blockchains: sync.Map{},
			//readyC: struct {}<-,
		}

		err := chainMaker.Init()

		if err != nil {
			t.Log(err)
		}

		serverList = append(serverList, chainMaker)
	}

	return serverList
}
