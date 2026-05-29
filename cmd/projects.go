package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/config"
	"github.com/spf13/cobra"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "Manage projects",
	}
	cmd.AddCommand(
		newProjectsListCmd(),
		newProjectsCreateCmd(),
		newProjectsDeleteCmd(),
		newProjectsUseCmd(),
		newProjectsShowCmd(),
	)
	return cmd
}

func projectRow(v any) []string {
	p := v.(api.Project)
	created := ""
	if !p.CreatedAt.IsZero() {
		created = p.CreatedAt.Format("2006-01-02")
	}
	return []string{p.ID, p.Name, created}
}

func newProjectsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.ListProjects(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "NAME", "CREATED"}, projectRow)
		},
	}
}

func newProjectsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			p, err := a.Client.CreateProject(cmd.Context(), a.Team, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*p, []string{"ID", "NAME"}, func(v any) []string {
				x := v.(api.Project)
				return []string{x.ID, x.Name}
			})
		},
	}
}

func newProjectsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete a project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.DeleteProject(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			return printAction(a, actionResult{ID: args[0], Status: "deleted"},
				fmt.Sprintf("Project %s deleted.", args[0]))
		},
	}
}

func newProjectsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			p, err := a.Client.GetProject(cmd.Context(), a.Team, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*p, []string{"ID", "NAME"}, func(v any) []string {
				x := v.(api.Project)
				return []string{x.ID, x.Name}
			})
		},
	}
}

func newProjectsUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <id>",
		Short: "Set the default project for this profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof, err := config.Load(profileName)
			if err != nil {
				return err
			}
			prof.DefaultProject = args[0]
			if err := config.Save(prof); err != nil {
				return err
			}
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			return printAction(a, actionResult{ID: args[0], Status: "default"},
				fmt.Sprintf("Default project set to %s.", args[0]))
		},
	}
}
