/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

func enableOrDisableGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas",
		Short: "enable or disable gas feature",
		Long:  "enable or disable gas feature",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			// required public key mode
			if sdk.AuthTypeToStringMap[cc.GetAuthType()] != protocol.Public {
				return errors.New("chainmaker must be Public Key Mode")
			}

			//// 2.Enable or disable gas feature
			var adminKeys []string
			if adminKeyFilePaths != "" {
				adminKeys = strings.Split(adminKeyFilePaths, ",")
			}
			if len(adminKeys) == 0 {
				return errors.New("admin key list is empty")
			}

			chainConfig, err := cc.GetChainConfig()
			if err != nil {
				return err
			}
			isGasEnabled := false
			if chainConfig.AccountConfig != nil {
				isGasEnabled = chainConfig.AccountConfig.EnableGas
			}

			if (gasEnable && !isGasEnabled) || (!gasEnable && isGasEnabled) {
				payload, err := cc.CreateChainConfigEnableOrDisableGasPayload()
				if err != nil {
					return err
				}
				endorsers := make([]*common.EndorsementEntry, len(adminKeys))
				for i := range adminKeys {
					var e *common.EndorsementEntry
					var err error
					e, err = sdkutils.MakePkEndorserWithPath(
						adminKeys[i],
						crypto.HashAlgoMap[cc.GetHashType()],
						"",
						payload,
					)
					if err != nil {
						return err
					}
					endorsers[i] = e
				}

				resp, err := cc.SendChainConfigUpdateRequest(payload, endorsers, -1, syncResult)
				if err != nil {
					return err
				}

				err = util.CheckProposalRequestResp(resp, false)
				if err != nil {
					return err
				}
			}

			output, err := prettyjson.Marshal("OK")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagSyncResult,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagGasEnable,
	})
	return cmd
}
