/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package parallel

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/spf13/cobra"
)

func createContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   createContractStr,
		Short: "Create Contract",
		Long:  "Create Contract",
		RunE: func(_ *cobra.Command, _ []string) error {
			return parallel(createContractStr)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&wasmPath, "wasm-path", "w", "../wasm/counter-go.wasm", "specify wasm path")
	flags.Int32VarP(&runTime, "run-time", "m", int32(commonPb.RuntimeType_GASM), "specify run time")

	return cmd
}
