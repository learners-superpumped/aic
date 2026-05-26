package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/auth"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/config"
	"github.com/spf13/cobra"
)

// ensureDefaultTeam returns the team id to set as the profile default for a
// freshly logged-in user. When the user has no teams it explicitly creates a
// "personal" team (the backend never auto-creates). It returns ("", false, nil)
// when the user already has a default team set and need not change it.
func ensureDefaultTeam(ctx context.Context, client *api.Client, currentDefault string) (teamID string, created bool, err error) {
	teams, err := client.ListTeams(ctx)
	if err != nil {
		return "", false, err
	}
	if len(teams) == 0 {
		t, err := client.CreateTeam(ctx, "personal")
		if err != nil {
			return "", false, err
		}
		return t.ID, true, nil
	}
	if currentDefault == "" {
		return teams[0].ID, false, nil
	}
	return "", false, nil
}

func newAuthCmds() []*cobra.Command {
	return []*cobra.Command{
		newLoginCmd(),
		newLogoutCmd(),
		newWhoamiCmd(),
		newConfigureCmd(),
	}
}

func newLoginCmd() *cobra.Command {
	var headless bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via your browser and store credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof, loadErr := config.Load(profileName)
			if prof == nil || prof.Issuer == "" || prof.ClientID == "" {
				if loadErr != nil {
					return fmt.Errorf("could not read configuration: %w (run `aic configure --issuer <url> --client-id <id>` if not yet configured)", loadErr)
				}
				return fmt.Errorf("OIDC issuer/client not configured: run `aic configure --issuer <url> --client-id <id>`")
			}

			oc, err := auth.Discover(cmd.Context(), prof.Issuer, prof.ClientID, prof.AudienceScope)
			if err != nil {
				return err
			}

			var ts *auth.TokenSet
			if headless {
				ts, err = auth.DeviceLogin(cmd.Context(), oc, func(uri, code string) {
					fmt.Printf("To sign in, visit %s and enter code: %s\n", uri, code)
				})
			} else {
				ts, err = auth.LoopbackLogin(cmd.Context(), oc, auth.OpenBrowser)
			}
			if err != nil {
				return err
			}

			prof.AccessToken = ts.AccessToken
			prof.RefreshToken = ts.RefreshToken
			prof.IDToken = ts.IDToken
			prof.ExpiresAt = ts.Expiry
			if err := config.Save(prof); err != nil {
				return err
			}

			// Bootstrap: ensure the user has a team to work in. The backend never
			// auto-creates teams; on first login the CLI explicitly creates one.
			// Auth already succeeded and credentials are saved, so a failure here
			// is demoted to a warning rather than failing the login.
			endpoint := prof.APIEndpoint
			if endpoint == "" {
				endpoint = defaultEndpoint
			}
			client := api.New(endpoint, ts.AccessToken)
			teamID, created, berr := ensureDefaultTeam(cmd.Context(), client, prof.Team)
			if berr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not set up your team (run `aic teams create <name>`): %v\n", berr)
			} else if teamID != "" {
				prof.Team = teamID
				if err := config.Save(prof); err != nil {
					return err
				}
				if created {
					fmt.Printf("Created your personal team %s.\n", teamID)
				}
			}
			fmt.Println("Login successful. Credentials saved.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&headless, "headless", false, "use the device code flow (no local browser)")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials for a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			if err := config.Delete(profileName); err != nil {
				return err
			}
			fmt.Println("Logged out.")
			return nil
		},
	}
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the currently authenticated identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof, err := config.Load(profileName)
			if err != nil {
				return err
			}
			if prof.IDToken == "" {
				return fmt.Errorf("not logged in: run `aic login`")
			}
			sub, email, err := auth.ParseIDTokenClaims(prof.IDToken)
			if err != nil {
				return err
			}
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			return a.Out.Print(
				map[string]string{"user_id": sub, "email": email},
				[]string{"USER ID", "EMAIL"},
				func(v any) []string {
					m := v.(map[string]string)
					return []string{m["user_id"], m["email"]}
				},
			)
		},
	}
}

func newConfigureCmd() *cobra.Command {
	var endpoint, output, issuer, clientID, audienceScope string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Set CLI configuration (API endpoint, output format)",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof, _ := config.Load(profileName)
			if prof == nil {
				prof = &config.Profile{Name: profileName}
			}
			if endpoint != "" {
				prof.APIEndpoint = endpoint
			}
			if output != "" {
				if err := validateOutputFormat(output); err != nil {
					return err
				}
				prof.Output = output
			}
			if issuer != "" {
				prof.Issuer = issuer
			}
			if clientID != "" {
				prof.ClientID = clientID
			}
			if cmd.Flags().Changed("audience-scope") {
				prof.AudienceScope = audienceScope
			}
			if err := config.Save(prof); err != nil {
				return err
			}
			fmt.Println("Configuration saved.")
			return nil
		},
	}
	cmd.Flags().StringVar(&endpoint, "api-endpoint", "", "backend API endpoint URL")
	cmd.Flags().StringVar(&output, "output-format", "", "default output format: table|json|yaml")
	cmd.Flags().StringVar(&issuer, "issuer", "", "OIDC issuer URL")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OIDC client id for the CLI")
	cmd.Flags().StringVar(&audienceScope, "audience-scope", "", "extra OIDC scope to request the API audience (provider-specific)")
	return cmd
}
