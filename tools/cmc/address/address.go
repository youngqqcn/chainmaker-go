// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package address

import (
	"fmt"
	"io/ioutil"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

func newPK2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pk-to-addr [public key file path / pem string]",
		Short: "get address from public key file or pem string",
		Long:  "get address from public key file or pem string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var keyPemStr string
			var isFile = utils.Exists(args[0])
			if isFile {
				keyPem, err := ioutil.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read key file failed, %s", err)
				}
				keyPemStr = string(keyPem)
			} else {
				keyPemStr = args[0]
			}

			var addr string
			var err error
			switch addressType {
			case addressTypeZXL:
				addr, err = sdk.GetZXAddressFromPKPEM(keyPemStr)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported address type %s", addressType)
			}

			output, err := prettyjson.Marshal(addr)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagAddressType})
	return cmd
}

func newHex2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hex-to-addr [hex string]",
		Short: "get address from hex string",
		Long:  "get address from hex string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var addr string
			var err error
			switch addressType {
			case addressTypeZXL:
				addr, err = sdk.GetZXAddressFromPKHex(args[0])
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported address type %s", addressType)
			}

			output, err := prettyjson.Marshal(addr)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagAddressType})
	return cmd
}

func newCert2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert-to-addr [hex string]",
		Short: "get address from cert file or pem string",
		Long:  "get address from cert file or pem string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var addr string
			var err error
			var isFile = utils.Exists(args[0])
			switch addressType {
			case addressTypeZXL:
				if isFile {
					addr, err = sdk.GetZXAddressFromCertPath(args[0])
					if err != nil {
						return err
					}
				} else {
					addr, err = sdk.GetZXAddressFromCertPEM(args[0])
					if err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported address type %s", addressType)
			}

			output, err := prettyjson.Marshal(addr)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagAddressType})
	return cmd
}
