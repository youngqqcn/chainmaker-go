package parallel

import "github.com/spf13/cobra"

func queryCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query",
		Long:  "Query",
		RunE: func(_ *cobra.Command, _ []string) error {
			return parallel("query")
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pairsString, "pairs", "a", "[{\"key\":\"key\",\"value\":\"counter1\",\"unique\":false}]", "specify pairs")
	flags.StringVarP(&pairsFile, "pairs-file", "A", "", "specify pairs file, if used, set --pairs=\"\"")
	flags.StringVarP(&method, "method", "m", "increase", "specify contract method")

	return cmd
}
