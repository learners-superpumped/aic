package cmd

import "github.com/spf13/cobra"

func newMailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Manage outbound mail (SES identities, inboxes, send)",
	}
	cmd.AddCommand(newMailDomainsCmd())
	return cmd
}
