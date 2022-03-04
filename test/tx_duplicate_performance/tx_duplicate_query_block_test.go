/*
   Created by guoxin in 2022/3/3 11:09 AM
*/
package tx_duplicate_performance

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"testing"
)

func TestQueryBlockSendTx(t *testing.T) {
	client, err := sdk.NewChainClient(
		sdk.WithConfPath("./sdk_config_solo.yml"),
		sdk.WithChainClientChainId("chain1"),
		//sdk.WithRPCClientConfig(&sdk.RPCClientConfig{4 * 1024 * 1024}),
	)
	if err != nil {
		panic(err)
	}

	var height uint64
	block, err := client.GetLastBlock(true)
	if err != nil {
		panic(err)
	}
	height = block.GetBlock().Header.BlockHeight
	ids := utils.GetTxIds(block.GetBlock().Txs)
	for _, id := range ids {
		res, err := client.InvokeContract("T", "P", id, []*common.KeyValuePair{
			{
				Key:   "method",
				Value: []byte("Base"),
			},
			{
				Key:   "file_name",
				Value: []byte("sssssfeifesssssfeifeitestsssssfeifeitestsssssfeifeitestsssssfeifeitestsssssfeifeitestitest"),
			},
			{
				Key:   "file_hash",
				Value: []byte("abcdqwrtasdf"),
			},
			{
				Key:   "time",
				Value: []byte("6543234"),
			},
		}, 10000, true)
		if err != nil {
			t.Logf("InvokeContract error %v", err)
		}

		if res.Code != common.TxStatusCode_SUCCESS {
			t.Logf("[ERROR] invoke contract failed, [code:%d]/[msg:%s]/[txId:%s]\n", res.Code, res.Message, id)
		}
	}
	for _, id := range ids {
		txId, err := client.GetTxByTxId(id)
		if err != nil {
			t.Logf("GetTxByTxId error %v", err)
		}
		if txId.BlockHeight != height {
			t.Logf("BlockHeight not equal got:%v, want:%v", txId.BlockHeight, height)
		}
	}

	//const (
	//	start1 = 1
	//)
	//var (
	//	start uint64 = start1
	//	end   uint64 = 0
	//)
	//
	//var first []string
	//storeNTxIds(10000, first, func() []string {
	//	block, err := client.GetBlockByHeight(start, true)
	//	if err != nil {
	//		panic(err)
	//	}
	//	start++
	//	return utils.GetTxIds(block.GetBlock().Txs)
	//})
	//t.Logf("size: %v, height: %v~%v", first, start1, start)
	//storeNTxIds(10000, first, func() []string {
	//	var block *common.BlockInfo
	//	if end == 0 {
	//		block, err = client.GetLastBlock(true)
	//		if err != nil {
	//			panic(err)
	//		}
	//		end = block.GetBlock().Header.BlockHeight
	//		return utils.GetTxIds(block.GetBlock().Txs)
	//	} else {
	//		block, err = client.GetBlockByHeight(start, true)
	//		if err != nil {
	//			panic(err)
	//		}
	//		end--
	//		return utils.GetTxIds(block.GetBlock().Txs)
	//	}
	//})

}

func storeNTxIds(n int, txIds []string, f func() []string) {
	for {
		ids := f()
		txIds = append(txIds, ids...)
		if len(txIds) >= n {
			return
		}
	}
}
