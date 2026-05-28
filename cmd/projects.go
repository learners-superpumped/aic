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
			return a.Out.Print(items, []string{"ID", "NAME", "CREATED"}, func(v any) []string {
				p := v.(api.Project)
				return []string{p.ID, p.Name, p.CreatedAt.Format("2006-01-02")}
			})
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
			fmt.Printf("Project %s deleted.\n", args[0])
			return nil
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
			fmt.Printf("Default project set to %s.\n", args[0])
			return nil
		},
	}
}
