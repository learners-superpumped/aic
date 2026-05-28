package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/app"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/auth"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/config"
	"github.com/spf13/cobra"
)

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

func resolveTeam(flag, def string) string {
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
	defaultTeam    string
	teamFlag       string
	refreshFn      func(context.Context) (*api.Tokens, error)
	onRefresh      func(*api.Tokens)
}

func buildApp(a buildAppArgs) (*app.App, error) {
	renderer, err := app.NewRenderer(a.output, os.Stdout)
	if err != nil {
		return nil, err
	}
	client := api.New(a.apiEndpoint, a.token)
	if a.refreshFn != nil {
		client = client.WithRefresh(a.refreshFn, a.onRefresh)
	}
	return &app.App{
		Client:  client,
		Project: resolveProject(a.projectFlag, a.defaultProject),
		Team:    resolveTeam(a.teamFlag, a.defaultTeam),
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
	root.PersistentFlags().StringP("team", "t", "", "target team (overrides the default team)")
	root.PersistentFlags().StringP("output", "o", "table", "output format: table|json|yaml")
	root.PersistentFlags().String("profile", "default", "credentials profile to use")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if commandsSkippingApp[cmd.Name()] {
			return nil
		}
		profileName, _ := cmd.Flags().GetString("profile")
		projectFlag, _ := cmd.Flags().GetString("project")
		teamFlag, _ := cmd.Flags().GetString("team")
		outFlag, _ := cmd.Flags().GetString("output")

		prof := config.LoadOrDefault(profileName)
		endpoint := prof.APIEndpoint
		output := outFlag
		if !cmd.Flags().Changed("output") && prof.Output != "" {
			output = prof.Output
		}

		onRefresh := func(tok *api.Tokens) {
			prof.AccessToken = tok.AccessToken
			if tok.RefreshToken != "" {
				prof.RefreshToken = tok.RefreshToken
			}
			prof.ExpiresAt = tok.ExpiresAt
			if err := config.Save(prof); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not persist refreshed token: %v\n", err)
			}
		}

		var refreshFn func(context.Context) (*api.Tokens, error)
		if prof.RefreshToken != "" {
			refreshFn = func(ctx context.Context) (*api.Tokens, error) {
				oc, err := auth.Discover(ctx, prof.Issuer, prof.ClientID, prof.AudienceScope)
				if err != nil {
					return nil, err
				}
				ts, err := auth.RefreshTokens(ctx, oc, prof.RefreshToken)
				if err != nil {
					return nil, err
				}
				return &api.Tokens{AccessToken: ts.AccessToken, RefreshToken: ts.RefreshToken, ExpiresAt: ts.Expiry}, nil
			}
		}

		a, err := buildApp(buildAppArgs{
			profileName:    profileName,
			output:         output,
			apiEndpoint:    endpoint,
			token:          prof.AccessToken,
			defaultProject: prof.DefaultProject,
			projectFlag:    projectFlag,
			defaultTeam:    prof.Team,
			teamFlag:       teamFlag,
			refreshFn:      refreshFn,
			onRefresh:      onRefresh,
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

	root.AddCommand(newTeamsCmd())
	root.AddCommand(newProjectsCmd())
	root.AddCommand(newDomainsCmd())
	root.AddCommand(newInboxesCmd())
	root.AddCommand(newMessagesCmd())
	root.AddCommand(newBillingCmd())

	return root
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
