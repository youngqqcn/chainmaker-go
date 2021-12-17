// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package zxl

import (
	"github.com/spf13/cobra"
)

func NewZXLCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zxl",
		Short: "zhi xin lian command",
		Long:  "zhi xin lian command",
	}

	cmd.AddCommand(newPK2AddrCMD())
	cmd.AddCommand(newHex2AddrCMD())
	cmd.AddCommand(newCert2AddrCMD())

	return cmd
}
