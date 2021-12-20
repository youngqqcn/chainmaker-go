// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package address

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagAddressType = "address-type"

	// address types
	addressTypeZXL = "zxl"
)

var (
	addressType string
)

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}

	flags.StringVar(&addressType, flagAddressType, "zxl", `The type of address obtained. 
eg. --address-type=zxl`)
}

func NewAddressCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "address parse command",
		Long:  "address parse command",
	}

	cmd.AddCommand(newPK2AddrCMD())
	cmd.AddCommand(newHex2AddrCMD())
	cmd.AddCommand(newCert2AddrCMD())

	return cmd
}
