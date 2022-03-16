// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package gas

import (
	"errors"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

func newSetGasAdminCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-admin [address]",
		Short: "set gas admin, set self as a admin if [address] not set",
		Long:  "set gas admin, set self as a admin if [address] not set",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			//// 2.Set gas admin
			var adminKeys []string
			if adminKeyFilePaths != "" {
				adminKeys = strings.Split(adminKeyFilePaths, ",")
			}
			if len(adminKeys) == 0 {
				return errors.New("admin key list is empty")
			}

			var adminAddr string
			if len(args) == 0 {
				pk, err := cc.GetPublicKey().String()
				if err != nil {
					return err
				}
				adminAddr, err = sdk.GetZXAddressFromPKPEM(pk)
				if err != nil {
					return err
				}
			} else {
				adminAddr = args[0]
			}

			payload, err := cc.CreateSetGasAdminPayload(adminAddr)
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

			resp, err := cc.SendGasManageRequest(payload, endorsers, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
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
		flagSdkConfPath,
	})
	return cmd
}

func newGetGasAdminCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-admin",
		Short: "get gas admin",
		Long:  "get gas admin",
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

			//// 2.Get gas admin
			addr, err := cc.GetGasAdmin()
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(addr)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func newRechargeGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recharge",
		Short: "recharge gas for account",
		Long:  "recharge gas for account",
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

			//// 2.Recharge gas
			rechargeGasList := []*syscontract.RechargeGas{
				{
					Address:   address,
					GasAmount: amount,
				},
			}
			payload, err := cc.CreateRechargeGasPayload(rechargeGasList)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagAddress, flagAmount,
	})
	return cmd
}

func newGetGasBalanceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-balance",
		Short: "get gas balance of account, get self balance if --address not set",
		Long:  "get gas balance of account, get self balance if --address not set",
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

			//// 2.Get gas balance
			if address == "" {
				pk, err := cc.GetPublicKey().String()
				if err != nil {
					return err
				}
				address, err = sdk.GetZXAddressFromPKPEM(pk)
				if err != nil {
					return err
				}
			}

			balance, err := cc.GetGasBalance(address)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(balance)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAddress,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func newRefundGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund",
		Short: "refund gas for account",
		Long:  "refund gas for account",
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

			//// 2.Refund gas
			payload, err := cc.CreateRefundGasPayload(address, amount)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagAddress, flagAmount,
	})
	return cmd
}

func newFrozenGasAccountCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen",
		Short: "frozen gas account",
		Long:  "frozen gas account",
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

			//// 2.Frozen gas account
			payload, err := cc.CreateFrozenGasAccountPayload(address)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagAddress,
	})
	return cmd
}

func newUnfrozenGasAccountCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfrozen",
		Short: "unfrozen gas account",
		Long:  "unfrozen gas account",
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

			//// 2.Unfrozen gas account
			payload, err := cc.CreateUnfrozenGasAccountPayload(address)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagAddress,
	})
	return cmd
}

func newGetGasAccountStatusCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-status",
		Short: "get gas account status, get self gas account status if --address not set",
		Long:  "get gas account status, get self gas account status if --address not set",
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

			//// 2.Get gas account status
			if address == "" {
				pk, err := cc.GetPublicKey().String()
				if err != nil {
					return err
				}
				address, err = sdk.GetZXAddressFromPKPEM(pk)
				if err != nil {
					return err
				}
			}

			status, err := cc.GetGasAccountStatus(address)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(status)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAddress,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}
