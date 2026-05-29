package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newTeamsInviteCmd() *cobra.Command {
	var role string
	cmd := &cobra.Command{
		Use:   "invite <email>",
		Short: "Invite a user to the current team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			inv, err := a.Client.CreateInvite(cmd.Context(), a.Team, args[0], role)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Invite sent to %s (id=%s, expires %s).\n",
				inv.Email, inv.ID, inv.ExpiresAt.UTC().Format("2006-01-02"))
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "member", "Role to grant: 'owner' or 'member'")
	return cmd
}

func newTeamsInvitesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "invites", Short: "Manage pending invites"}
	cmd.AddCommand(
		newTeamsInvitesListCmd(),
		newTeamsInvitesRevokeCmd(),
		newTeamsInvitesResendCmd(),
	)
	return cmd
}

func newTeamsInvitesListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List pending invites",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.ListInvites(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "EMAIL", "ROLE", "EXPIRES"}, func(v any) []string {
				i := v.(api.Invite)
				return []string{i.ID, i.Email, i.Role, i.ExpiresAt.UTC().Format("2006-01-02")}
			})
		},
	}
}

func newTeamsInvitesRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use: "revoke <inviteId>", Short: "Revoke a pending invite", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.RevokeInvite(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Invite revoked.")
			return nil
		},
	}
}

func newTeamsInvitesResendCmd() *cobra.Command {
	return &cobra.Command{
		Use: "resend <inviteId>", Short: "Rotate token and resend the invite email", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			inv, err := a.Client.ResendInvite(cmd.Context(), a.Team, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"New invite sent to %s (id=%s, expires %s).\n",
				inv.Email, inv.ID, inv.ExpiresAt.UTC().Format("2006-01-02"))
			return nil
		},
	}
}
