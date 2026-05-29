package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInvitesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "invites", Short: "Accept or preview team invites"}
	cmd.AddCommand(newInvitesAcceptCmd(), newInvitesShowCmd())
	return cmd
}

func newInvitesAcceptCmd() *cobra.Command {
	return &cobra.Command{
		Use: "accept <token>", Short: "Accept a team invite", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			t, err := a.Client.AcceptInvite(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Joined team %s (id=%s) as %s.\n", t.Name, t.ID, t.Role)
			return nil
		},
	}
}

func newInvitesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use: "show <token>", Short: "Preview a team invite", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			p, err := a.Client.PreviewInvite(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Team: %s\nRole: %s\nInvited by: %s\n", p.TeamName, p.Role, p.InvitedByEmail)
			return nil
		},
	}
}
