package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rsshub-gateway",
		Short: "rsshub-gateway is a gateway for RSSHub.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd(version, buildTime, gitCommit))
	cmd.AddCommand(newServeCmd())
	return cmd
}

func Execute(version string, buildTime string, gitCommit string) error {
	if err := newRootCmd(version, buildTime, gitCommit).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
