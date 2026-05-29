package cmd

import (
	"fmt"
	"strings"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newMailInboxesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "inboxes", Aliases: []string{"inbox"}, Short: "Manage sending addresses (inboxes)"}
	cmd.AddCommand(
		newMailInboxesCreateCmd(),
		newMailInboxesListCmd(),
		newMailInboxesShowCmd(),
		newMailInboxesDeleteCmd(),
	)
	return cmd
}

func inboxRows() ([]string, func(any) []string) {
	return []string{"ADDRESS", "DISPLAY", "CREATED"}, func(v any) []string {
		x := v.(api.MailInbox)
		return []string{x.Address, x.DisplayName, x.CreatedAt}
	}
}

func splitInboxAddress(addr string) (local, domain string, err error) {
	i := strings.LastIndexByte(addr, '@')
	if i <= 0 || i == len(addr)-1 {
		return "", "", fmt.Errorf("invalid address %q", addr)
	}
	return addr[:i], addr[i+1:], nil
}

func newMailInboxesCreateCmd() *cobra.Command {
	var display string
	cmd := &cobra.Command{
		Use:   "create <address>",
		Short: "Register a sending address on an email-enabled domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			local, domain, err := splitInboxAddress(args[0])
			if err != nil {
				return err
			}
			in, err := a.Client.CreateMailInbox(cmd.Context(), a.Team, a.Project, domain, local, display)
			if err != nil {
				return err
			}
			cols, row := inboxRows()
			return a.Out.Print(*in, cols, row)
		},
	}
	cmd.Flags().StringVar(&display, "display-name", "", "display name shown in the From header")
	return cmd
}

func newMailInboxesListCmd() *cobra.Command {
	var domain string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List inboxes for a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if domain == "" {
				return fmt.Errorf("--domain is required")
			}
			items, err := a.Client.ListMailInboxes(cmd.Context(), a.Team, a.Project, domain)
			if err != nil {
				return err
			}
			cols, row := inboxRows()
			return a.Out.Print(items, cols, row)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "domain to list inboxes for (required)")
	return cmd
}

func newMailInboxesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <address>",
		Short: "Show an inbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			local, domain, err := splitInboxAddress(args[0])
			if err != nil {
				return err
			}
			in, err := a.Client.ShowMailInbox(cmd.Context(), a.Team, a.Project, domain, local)
			if err != nil {
				return err
			}
			cols, row := inboxRows()
			return a.Out.Print(*in, cols, row)
		},
	}
}

func newMailInboxesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <address>",
		Aliases: []string{"rm"},
		Short:   "Delete an inbox",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			local, domain, err := splitInboxAddress(args[0])
			if err != nil {
				return err
			}
			if err := a.Client.DeleteMailInbox(cmd.Context(), a.Team, a.Project, domain, local); err != nil {
				return err
			}
			return printAction(a, actionResult{Name: args[0], Status: "deleted"},
				fmt.Sprintf("Inbox %s deleted.", args[0]))
		},
	}
}
