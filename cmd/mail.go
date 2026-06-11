package cmd

import "github.com/spf13/cobra"

func newMailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Email sending and inboxes on your domains",
	}
	cmd.AddCommand(newMailDomainsCmd(), newMailInboxesCmd(), newMailSendCmd(), newMailMessagesCmd())
	return cmd
}
