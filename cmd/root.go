package cmd

import (
	"os"

	"github.com/learners-company/aic/internal/api"
	"github.com/learners-company/aic/internal/app"
	"github.com/learners-company/aic/internal/config"
	"github.com/spf13/cobra"
)

const defaultEndpoint = "https://api.aic.example.com"

// commandsSkippingApp lists commands that must NOT trigger credential/project
// resolution (they run before credentials exist).
var commandsSkippingApp = map[string]bool{
	"login":     true,
	"logout":    true,
	"configure": true,
	"help":      true,
}

func resolveProject(flag, def string) string {
	if flag != "" {
		return flag
	}
	return def
}

type buildAppArgs struct {
	profileName    string
	output         string
	apiEndpoint    string
	token          string
	defaultProject string
	projectFlag    string
}

func buildApp(a buildAppArgs) (*app.App, error) {
	renderer, err := app.NewRenderer(a.output, os.Stdout)
	if err != nil {
		return nil, err
	}
	return &app.App{
		Client:  api.New(a.apiEndpoint, a.token),
		Project: resolveProject(a.projectFlag, a.defaultProject),
		Out:     renderer,
	}, nil
}

// appFromCmd is the one-liner every command uses to get its runtime context.
func appFromCmd(cmd *cobra.Command) (*app.App, error) {
	return app.FromContext(cmd.Context())
}

// validateOutputFormat returns an error if format is not a supported renderer format.
func validateOutputFormat(format string) error {
	_, err := app.NewRenderer(format, os.Stdout)
	return err
}

// NewRootCmd builds the top-level `aic` command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "aic",
		Short:         "aic provisions projects, domains, and email inboxes on our service",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringP("project", "p", "", "target project (overrides the default project)")
	root.PersistentFlags().StringP("output", "o", "table", "output format: table|json|yaml")
	root.PersistentFlags().String("profile", "default", "credentials profile to use")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if commandsSkippingApp[cmd.Name()] {
			return nil
		}
		profileName, _ := cmd.Flags().GetString("profile")
		projectFlag, _ := cmd.Flags().GetString("project")
		outFlag, _ := cmd.Flags().GetString("output")

		prof, err := config.Load(profileName)
		if err != nil {
			return err
		}
		endpoint := prof.APIEndpoint
		if endpoint == "" {
			endpoint = defaultEndpoint
		}
		output := outFlag
		if !cmd.Flags().Changed("output") && prof.Output != "" {
			output = prof.Output
		}

		a, err := buildApp(buildAppArgs{
			profileName:    profileName,
			output:         output,
			apiEndpoint:    endpoint,
			token:          prof.AccessToken,
			defaultProject: prof.DefaultProject,
			projectFlag:    projectFlag,
		})
		if err != nil {
			return err
		}
		cmd.SetContext(app.NewContext(cmd.Context(), a))
		return nil
	}

	// Auth subcommands (login, logout, whoami, configure).
	// Later tasks add more commands here.
	for _, c := range newAuthCmds() {
		root.AddCommand(c)
	}

	root.AddCommand(newProjectsCmd())
	root.AddCommand(newDomainsCmd())
	root.AddCommand(newInboxesCmd())
	root.AddCommand(newMessagesCmd())

	return root
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
