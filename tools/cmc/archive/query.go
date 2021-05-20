/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package archive

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/spf13/cobra"

	"chainmaker.org/chainmaker-go/tools/cmc/archive/model"
	"chainmaker.org/chainmaker-sdk-go/pb/protogo/common"
	"chainmaker.org/chainmaker-sdk-go/pb/protogo/store"
)

func newQueryOffChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query off-chain blockchain data",
		Long:  "query off-chain blockchain data",
	}

	cmd.AddCommand(newQueryTxOffChainCMD())
	cmd.AddCommand(newQueryBlockOffChainCMD())
	cmd.AddCommand(newQueryArchivedHeightOffChainCMD())

	return cmd
}

func newQueryTxOffChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [txid]",
		Short: "query off-chain tx by txid",
		Long:  "query off-chain tx by txid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := createChainClient(adminKeyFilePaths, adminCrtFilePaths, chainId)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Database
			db, err := initDb()
			if err != nil {
				return err
			}

			//// 3.Query tx off-chain.
			var txInfo *common.TransactionInfo
			var output []byte
			blkHeight, err := cc.GetBlockHeightByTxId(args[0])
			if err != nil {
				return err
			}

			fmt.Println("blkHeight==", blkHeight)
			var bInfo model.BlockInfo
			err = db.Table(model.BlockInfoTableNameByBlockHeight(blkHeight)).Where("Fblock_height = ?", blkHeight).Find(&bInfo).Error
			if err != nil {
				return err
			}

			var blkWithRWSet store.BlockWithRWSet
			err = blkWithRWSet.Unmarshal(bInfo.BlockWithRWSet)
			if err != nil {
				return err
			}

			if blkWithRWSet.Block != nil {
				for idx, tx := range blkWithRWSet.Block.Txs {
					if tx.Header.TxId == args[0] {
						txInfo = &common.TransactionInfo{
							Transaction: tx,
							BlockHeight: uint64(blkWithRWSet.Block.Header.BlockHeight),
							BlockHash:   blkWithRWSet.Block.Header.BlockHash,
							TxIndex:     uint32(idx),
						}

						output, err = txInfo.Marshal()
						if err != nil {
							return err
						}
						break
					}
				}
			}

			if txInfo == nil {
				output, _ = json.MarshalIndent(map[string]string{"err": "tx not found in off-chain storage"}, "", "    ")
			} else {
				output, err = json.MarshalIndent(txInfo, "", "    ")
				if err != nil {
					return err
				}
			}

			fmt.Println(string(output))
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath, flagChainId, flagDbType, flagDbDest,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagChainId)
	cmd.MarkFlagRequired(flagDbType)
	cmd.MarkFlagRequired(flagDbDest)

	return cmd
}

func newQueryBlockOffChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "query off-chain block by height",
		Long:  "query off-chain block by height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			//// 1.Database
			db, err := initDb()
			if err != nil {
				return err
			}

			//// 2.Query block off-chain.
			var output []byte
			var bInfo model.BlockInfo
			err = db.Table(model.BlockInfoTableNameByBlockHeight(height)).Where(&model.BlockInfo{BlockHeight: height}).Find(&bInfo).Error
			if err != nil {
				return err
			}

			if reflect.DeepEqual(bInfo, model.BlockInfo{}) {
				output, _ = json.MarshalIndent(map[string]string{"err": "block not found in off-chain storage"}, "", "    ")
			} else {
				var blkWithRWSetOffChain store.BlockWithRWSet
				err = blkWithRWSetOffChain.Unmarshal(bInfo.BlockWithRWSet)
				if err != nil {
					return err
				}
				output, err = json.MarshalIndent(blkWithRWSetOffChain, "", "    ")
				if err != nil {
					return err
				}
			}

			fmt.Println(string(output))
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath, flagChainId, flagDbType, flagDbDest,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagChainId)
	cmd.MarkFlagRequired(flagDbType)
	cmd.MarkFlagRequired(flagDbDest)

	return cmd
}

func newQueryArchivedHeightOffChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archived-height",
		Short: "query off-chain archived height",
		Long:  "query off-chain archived height",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Database
			db, err := initDb()
			if err != nil {
				return err
			}

			//// 2.Query archived block height off-chain.
			archivedBlkHeightOffChain, err := model.GetArchivedBlockHeight(db)
			if err != nil {
				return err
			}

			output, err := json.MarshalIndent(map[string]int64{"archived_height": archivedBlkHeightOffChain}, "", "    ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath, flagChainId, flagDbType, flagDbDest,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagChainId)
	cmd.MarkFlagRequired(flagDbType)
	cmd.MarkFlagRequired(flagDbDest)

	return cmd
}
