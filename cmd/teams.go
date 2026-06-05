package cmd

import (
	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/config"
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
		newTeamsInviteCmd(),
		newTeamsInvitesCmd(),
		newTeamsMembersCmd(),
	)
	return cmd
}

func teamRows() ([]string, func(any) []string) {
	return []string{"ID", "NAME", "ROLE"}, func(v any) []string {
		t := v.(api.Team)
		return []string{t.ID, t.Name, t.Role}
	}
}

func newTeamsListCmd() *cobra.Command {
	var limit int
	var cursor string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List teams you belong to",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			page, err := a.Client.ListTeams(cmd.Context(), limit, cursor)
			if err != nil {
				return err
			}
			cols, row := teamRows()
			return printPage(cmd, a.Out, page, cols, row)
		},
	}
	addPaginationFlags(cmd, &limit, &cursor)
	return cmd
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
			cols, row := teamRows()
			return a.Out.Print(*t, cols, row)
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
			t, err := a.Client.GetTeam(cmd.Context(), args[0])
			if err != nil {
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
			cols, row := teamRows()
			return a.Out.Print(*t, cols, row)
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
			t, err := a.Client.GetTeam(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			cols, row := teamRows()
			return a.Out.Print(*t, cols, row)
		},
	}
}
