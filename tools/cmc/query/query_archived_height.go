// Copyright (C) BABEC. All rights reserved.
// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

func newQueryArchivedHeightOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archived-height",
		Short: "query on-chain archived height",
		Long:  "query on-chain archived height",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQueryArchivedHeightOnChainCMD()
		},
	}

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagChainId,
	})
	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash,
	})
	return cmd
}

// runQueryArchivedHeightOnChainCMD `query archived height` command implementation
func runQueryArchivedHeightOnChainCMD() error {
	//// 1.Chain Client
	cc, err := sdk.NewChainClient(
		sdk.WithConfPath(sdkConfPath),
		sdk.WithChainClientChainId(chainId),
	)
	if err != nil {
		return err
	}
	defer cc.Stop()
	if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
		return err
	}

	//// 2.Query archived height
	archivedBlkHeight, err := cc.GetArchivedBlockHeight()
	if err != nil {
		return err
	}

	output, err := prettyjson.Marshal(map[string]uint64{"archived_height": archivedBlkHeight})
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
