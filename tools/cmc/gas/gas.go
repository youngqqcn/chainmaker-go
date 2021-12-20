// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package gas

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sdkConfPath       string
	chainId           string
	syncResult        bool
	adminKeyFilePaths string
	address           string
	amount            int64
)

const (
	flagSdkConfPath       = "sdk-conf-path"
	flagChainId           = "chain-id"
	flagSyncResult        = "sync-result"
	flagAdminKeyFilePaths = "admin-key-file-paths"
	flagAddress           = "address"
	flagAmount            = "amount"
)

func NewGasManageCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas",
		Short: "gas management",
		Long:  "gas management",
	}

	cmd.AddCommand(newSetGasAdminCMD())
	cmd.AddCommand(newGetGasAdminCMD())
	cmd.AddCommand(newRechargeGasCMD())
	cmd.AddCommand(newGetGasBalanceCMD())
	cmd.AddCommand(newRefundGasCMD())
	cmd.AddCommand(newFrozenGasAccountCMD())
	cmd.AddCommand(newUnfrozenGasAccountCMD())
	cmd.AddCommand(newGetGasAccountStatusCMD())

	return cmd
}

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}

	flags.StringVar(&chainId, flagChainId, "", "Chain ID")
	flags.StringVar(&sdkConfPath, flagSdkConfPath, "", "specify sdk config path")
	flags.BoolVar(&syncResult, flagSyncResult, false, "whether wait the result of the transaction, default false")
	flags.StringVar(&adminKeyFilePaths, flagAdminKeyFilePaths, "", "specify admin key file paths, use ',' to separate")
	flags.StringVar(&address, flagAddress, "", "address of account")
	flags.Int64Var(&amount, flagAmount, 0, "amount of gas")
}
