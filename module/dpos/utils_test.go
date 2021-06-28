/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dpos

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"chainmaker.org/chainmaker-go/mock"
	commonpb "chainmaker.org/chainmaker-go/pb/protogo/common"
	pbdpos "chainmaker.org/chainmaker-go/pb/protogo/dpos"
	"chainmaker.org/chainmaker-go/vm/native"
	"github.com/golang/protobuf/proto"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCandidateInfos(t *testing.T) {
	var tests = []*pbdpos.CandidateInfo{
		{"peer0", "100"},
		{"peer1", "100"},
		{"peer2", "100"},
		{"peer3", "0"},
		{"peer4", "300"},
		{"peer5", "500"},
	}
	sort.Sort(CandidateInfos(tests))
	require.Equal(t, tests[0].Weight, "500")
	require.Equal(t, tests[1].Weight, "300")
	require.Equal(t, tests[2].Weight, "100")
	require.Equal(t, tests[3].Weight, "100")
	require.Equal(t, tests[4].Weight, "100")
	require.Equal(t, tests[5].Weight, "0")
	require.Equal(t, tests[0].PeerID, "peer5")
	require.Equal(t, tests[1].PeerID, "peer4")
	require.Equal(t, tests[2].PeerID, "peer0")
	require.Equal(t, tests[3].PeerID, "peer1")
	require.Equal(t, tests[4].PeerID, "peer2")
	require.Equal(t, tests[5].PeerID, "peer3")
}

func TestValidatorsElection(t *testing.T) {
	var tests = []*pbdpos.CandidateInfo{
		{"peer0", "100"},
		{"peer1", "100"},
		{"peer2", "100"},
		{"peer3", "0"},
		{"peer4", "300"},
		{"peer5", "500"},
		{"peer6", "200"},
		{"peer7", "400"},
		{"peer8", "550"},
		{"peer9", "250"},
		{"peer10", "150"},
		{"peer11", "600"},
		{"peer12", "601"},
		{"peer13", "660"},
		{"peer14", "1000"},
	}
	seed := make([]byte, 32)
	rand.Read(seed)
	validators, err := ValidatorsElection(tests, 0, seed, false)
	require.NotNil(t, err)
	require.Nil(t, validators)
	validators, err = ValidatorsElection(tests, len(tests)+1, seed, false)
	require.NotNil(t, err)
	require.Nil(t, validators)
	validators, err = ValidatorsElection(tests, len(tests), seed, false)
	require.Equal(t, len(validators), len(tests))
	require.Nil(t, err)
	validators, err = ValidatorsElection(tests, 5, seed, false)
	require.Nil(t, err)
	require.Equal(t, len(validators), 5)
	for i := 0; i < len(validators); i++ {
		fmt.Printf("%v -> %s -> %s \n", i+1, validators[i].PeerID, validators[i].Weight)
	}
	for i := 0; i < 10; i++ {
		fmt.Println("----------------------------------")
		validators, err = ValidatorsElection(tests, i+1, seed, true)
		require.Nil(t, err)
		require.Equal(t, len(validators), i+1)
		for i := 0; i < len(validators); i++ {
			fmt.Printf("%v -> %s -> %s \n", i+1, validators[i].PeerID, validators[i].Weight)
		}
	}
	fmt.Println("----------------------------------")
	validators, err = ValidatorsElection(tests, len(tests)-1, seed, false)
	require.Nil(t, err)
	require.Equal(t, len(validators), len(tests)-1)
	var count = 0
	for i := 0; i < len(validators); i++ {
		peerID := validators[i].PeerID
		for j := 0; j < len(tests); j++ {
			if strings.EqualFold(peerID, tests[j].PeerID) {
				count++
				break
			}
		}
		fmt.Printf("%v -> %s -> %s \n", i+1, validators[i].PeerID, validators[i].Weight)
	}
	require.Equal(t, len(tests)-1, count)

	validators, err = ValidatorsElection(tests, 5, seed, true)
	require.Nil(t, err)
	require.Equal(t, len(validators), 5)
	for i := 0; i < 500; i++ {
		tmp, err := ValidatorsElection(tests, 5, seed, true)
		require.NoError(t, err)
		for i, v := range validators {
			if !strings.EqualFold(v.String(), tmp[i].String()) {
				fmt.Println("expect: ", validators)
				fmt.Println("actual: ", tmp)
				require.False(t, true)
			}
		}
		//require.EqualValues(t, validators, tmp)
	}
}

func TestRandPerm(t *testing.T) {
	for i := 0; i < 1000; i++ {
		rand.Seed(time.Now().Unix() + int64(i*20)) // 设置种子
		randSlice := rand.Perm(20)[:8]
		hasSeen := make(map[int]struct{}, len(randSlice))
		for _, v := range randSlice {
			if _, ok := hasSeen[v]; ok {
				require.False(t, ok, fmt.Sprintf("should not be repetition in randSlice"))
			} else {
				hasSeen[v] = struct{}{}
			}
		}
		fmt.Println(randSlice)
	}
}

func TestGetLatestEpochInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := mock.NewMockBlockchainStore(ctrl)
	mockStore.EXPECT().ReadObject(gomock.Any(), gomock.Any()).DoAndReturn(func(contractName string, key []byte) ([]byte, error) {
		epoch := &commonpb.Epoch{EpochID: 100, NextEpochCreateHeight: 990, ProposerVector: []string{
			"vector1", "vector2", "vector3", "vector4"}}
		return proto.Marshal(epoch)
	}).AnyTimes()
	epoch, err := GetLatestEpochInfo(mockStore)
	require.NoError(t, err)
	require.EqualValues(t, epoch.EpochID, 100)
	require.EqualValues(t, epoch.NextEpochCreateHeight, 990)
	require.EqualValues(t, epoch.ProposerVector, []string{
		"vector1", "vector2", "vector3", "vector4",
	})
}

func TestGetNodeIDsFromValidators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	name := commonpb.ContractName_SYSTEM_CONTRACT_DPOS_STAKE.String()
	nodeIDs := make(map[string]string)
	nodeIDs[name+string(native.ToNodeIDKey("val1"))] = "nodeId1"
	nodeIDs[name+string(native.ToNodeIDKey("val2"))] = "nodeId2"
	nodeIDs[name+string(native.ToNodeIDKey("val3"))] = "nodeId3"

	mockStore := mock.NewMockBlockchainStore(ctrl)
	mockStore.EXPECT().ReadObject(gomock.Any(), gomock.Any()).DoAndReturn(func(contractName string, key []byte) ([]byte, error) {
		val, exist := nodeIDs[contractName+string(key)]
		if exist {
			return []byte(val), nil
		}
		return nil, fmt.Errorf("not find key: %s in contract: %s", key, contractName)
	}).AnyTimes()
	ids, err := GetNodeIDsFromValidators(mockStore, []string{"val1", "val2", "val3"})
	require.NoError(t, err)
	require.EqualValues(t, ids, []string{"nodeId1", "nodeId2", "nodeId3"})
}
