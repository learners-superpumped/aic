package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/config"
	"github.com/spf13/cobra"
)

func newTeamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "teams",
		Aliases: []string{"team"},
		Short:   "Manage teams",
	}
	cmd.AddCommand(
		newTeamsListCmd(),
		newTeamsCreateCmd(),
		newTeamsSwitchCmd(),
		newTeamsShowCmd(),
	)
	return cmd
}

func newTeamsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List teams you belong to",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			items, err := a.Client.ListTeams(cmd.Context())
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "NAME", "ROLE"}, func(v any) []string {
				t := v.(api.Team)
				return []string{t.ID, t.Name, t.Role}
			})
		},
	}
}

func newTeamsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			t, err := a.Client.CreateTeam(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*t, []string{"ID", "NAME", "ROLE"}, func(v any) []string {
				x := v.(api.Team)
				return []string{x.ID, x.Name, x.Role}
			})
		},
	}
}

func newTeamsSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <id>",
		Short: "Set the default team for this profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			// Validate the team exists and the caller can see it.
			if _, err := a.Client.GetTeam(cmd.Context(), args[0]); err != nil {
				return err
			}
			profileName, _ := cmd.Flags().GetString("profile")
			prof, err := config.Load(profileName)
			if err != nil {
				return err
			}
			prof.Team = args[0]
			if err := config.Save(prof); err != nil {
				return err
			}
			fmt.Printf("Default team set to %s.\n", args[0])
			return nil
		},
	}
}

func newTeamsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current default team",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			fmt.Println(a.Team)
			return nil
		},
	}
}
