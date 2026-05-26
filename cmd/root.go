package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the top-level `aic` command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "aic",
		Short:         "aic provisions projects, domains, and email inboxes on our service",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringP("project", "p", "", "target project (overrides the default project)")
	root.PersistentFlags().StringP("output", "o", "table", "output format: table|json|yaml")
	root.PersistentFlags().String("profile", "default", "credentials profile to use")

	return root
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
