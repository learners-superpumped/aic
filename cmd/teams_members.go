package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newTeamsMembersCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "members", Short: "Manage team members"}
	cmd.AddCommand(newMembersListCmd(), newMembersRemoveCmd(), newMembersSetRoleCmd())
	return cmd
}

func newMembersListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List team members",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.ListMembers(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"USER_SUB", "EMAIL", "NAME", "ROLE", "JOINED"}, func(v any) []string {
				m := v.(api.Member)
				return []string{m.UserSub, m.Email, m.Name, m.Role, m.JoinedAt}
			})
		},
	}
}

func newMembersRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use: "remove <user-sub>", Short: "Remove a member from the team", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.RemoveMember(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Member removed.")
			return nil
		},
	}
}

func newMembersSetRoleCmd() *cobra.Command {
	return &cobra.Command{
		Use: "set-role <user-sub> <role>", Short: "Change a member's role", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			m, err := a.Client.SetMemberRole(cmd.Context(), a.Team, args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s is now %s.\n", m.UserSub, m.Role)
			return nil
		},
	}
}
