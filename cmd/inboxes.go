package cmd

import (
	"fmt"

	"github.com/learners-company/aic/internal/api"
	"github.com/spf13/cobra"
)

func newInboxesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inboxes",
		Aliases: []string{"inbox"},
		Short:   "Create and manage email inboxes",
	}
	cmd.AddCommand(
		newInboxesCreateCmd(),
		newInboxesListCmd(),
		newInboxesDeleteCmd(),
		newInboxesShowCmd(),
	)
	return cmd
}

func newInboxesCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <address>",
		Short: "Create an inbox on an owned domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			in, err := a.Client.CreateInbox(cmd.Context(), a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*in, []string{"ADDRESS", "STATUS"}, func(v any) []string {
				x := v.(api.Inbox)
				return []string{x.Address, x.Status}
			})
		},
	}
}

func newInboxesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List inboxes in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			items, err := a.Client.ListInboxes(cmd.Context(), a.Project)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ADDRESS", "STATUS"}, func(v any) []string {
				x := v.(api.Inbox)
				return []string{x.Address, x.Status}
			})
		},
	}
}

func newInboxesDeleteCmd() *cobra.Command {
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
			if err := a.Client.DeleteInbox(cmd.Context(), a.Project, args[0]); err != nil {
				return err
			}
			fmt.Printf("Inbox %s deleted.\n", args[0])
			return nil
		},
	}
}

func newInboxesShowCmd() *cobra.Command {
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
			in, err := a.Client.GetInbox(cmd.Context(), a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*in, []string{"ADDRESS", "STATUS"}, func(v any) []string {
				x := v.(api.Inbox)
				return []string{x.Address, x.Status}
			})
		},
	}
}
