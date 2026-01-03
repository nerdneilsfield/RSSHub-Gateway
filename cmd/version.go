package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "rsshub-gateway version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("rsshub-gateway")
			fmt.Println("Gateway for multi-instance RSSHub deployments.")
			fmt.Println("Author: dengqi935@gmail.com")
			fmt.Println("Github: https://github.com/nerdneilsfield/RSSHub-Gateway")
			fmt.Fprintf(cmd.OutOrStdout(), "rsshub-gateway: %s\n", version)
			fmt.Fprintf(cmd.OutOrStdout(), "buildTime: %s\n", buildTime)
			fmt.Fprintf(cmd.OutOrStdout(), "gitCommit: %s\n", gitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "goVersion: %s\n", runtime.Version())
		},
	}
}
