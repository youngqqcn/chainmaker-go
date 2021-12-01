package parallel

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/spf13/cobra"
)

func upgradeContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   upgradeContractStr,
		Short: "Upgrade Contract",
		Long:  "Upgrade Contract",
		RunE: func(_ *cobra.Command, _ []string) error {
			return parallel(upgradeContractStr)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&wasmPath, "wasm-path", "w", "../wasm/counter-go.wasm", "specify wasm path")
	flags.Int32VarP(&runTime, "run-time", "R", int32(commonPb.RuntimeType_GASM), "specify run time")
	flags.StringVarP(&version, "version", "v", "2.0.0", "specify contract version")

	return cmd
}
