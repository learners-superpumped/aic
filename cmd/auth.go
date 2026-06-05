package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/auth"
	"github.com/learners-superpumped/aic/internal/config"
	"github.com/spf13/cobra"
)

// ensureDefaultTeam returns the team id to set as the profile default for a
// freshly logged-in user. When the user has no teams it explicitly creates a
// "personal" team (the backend never auto-creates). It returns ("", false, nil)
// when the user already has a default team set and need not change it.
func ensureDefaultTeam(ctx context.Context, client *api.Client, currentDefault string) (teamID string, created bool, err error) {
	page, err := client.ListTeams(ctx, 1, "")
	if err != nil {
		return "", false, err
	}
	if len(page.Data) == 0 {
		t, err := client.CreateTeam(ctx, "personal")
		if err != nil {
			return "", false, err
		}
		return t.ID, true, nil
	}
	if currentDefault == "" {
		return page.Data[0].ID, false, nil
	}
	return "", false, nil
}

// newAuthCmd groups the identity commands under `aic auth` (login/logout/status),
// matching the `gh auth` / `gcloud auth` convention. `configure` is registered
// separately at the top level — it sets the service endpoint, not your identity.
func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Authenticate and inspect your account"}
	cmd.AddCommand(newLoginCmd(), newLogoutCmd(), newAuthStatusCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	var headless bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via your browser and store credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof := config.LoadOrDefault(profileName)

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
			client := api.New(prof.APIEndpoint, ts.AccessToken)
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

// newAuthStatusCmd shows the account snapshot: identity, current working context
// (team/project/credit), and the teams you belong to. Network lookups are
// best-effort — a failed call leaves its section empty rather than erroring.
func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show your account: identity, context, credit, and teams",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")
			prof, err := config.Load(profileName)
			if err != nil {
				return err
			}
			if prof.IDToken == "" {
				return fmt.Errorf("not logged in: run `aic auth login`")
			}
			sub, email, err := auth.ParseIDTokenClaims(prof.IDToken)
			if err != nil {
				return err
			}
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()

			st := api.AuthStatus{
				UserID: sub, Email: email, APIEndpoint: prof.APIEndpoint,
				DefaultTeamID: a.Team, DefaultProjectID: a.Project,
			}
			if teamsPage, terr := a.Client.ListTeams(ctx, 0, ""); terr == nil {
				st.Teams = teamsPage.Data
				for i := range teamsPage.Data {
					if teamsPage.Data[i].ID == a.Team {
						st.DefaultTeam = &teamsPage.Data[i]
					}
				}
			}
			if a.Team != "" {
				if bal, berr := a.Client.Balance(ctx, a.Team); berr == nil {
					st.BalanceUSD = bal.BalanceUSD
				}
				if a.Project != "" {
					if projsPage, perr := a.Client.ListProjects(ctx, a.Team, 0, ""); perr == nil {
						for i := range projsPage.Data {
							if projsPage.Data[i].ID == a.Project {
								st.DefaultProject = &projsPage.Data[i]
							}
						}
					}
				}
			}

			if a.Out.Format() != "table" {
				return a.Out.Print(st, nil, nil)
			}
			printAuthStatus(a.Out.Writer(), st)
			return nil
		},
	}
}

func printAuthStatus(w io.Writer, st api.AuthStatus) {
	fmt.Fprintf(w, "Identity\n")
	fmt.Fprintf(w, "  User    %s\n", st.UserID)
	if st.Email != "" {
		fmt.Fprintf(w, "  Email   %s\n", st.Email)
	}
	team := st.DefaultTeamID
	if st.DefaultTeam != nil {
		team = fmt.Sprintf("%s (%s) [%s]", st.DefaultTeam.ID, st.DefaultTeam.Name, st.DefaultTeam.Role)
	}
	project := st.DefaultProjectID
	if st.DefaultProject != nil {
		project = fmt.Sprintf("%s (%s)", st.DefaultProject.ID, st.DefaultProject.Name)
	}
	fmt.Fprintf(w, "\nContext\n")
	fmt.Fprintf(w, "  Team     %s\n", dashIfEmpty(team))
	fmt.Fprintf(w, "  Project  %s\n", dashIfEmpty(project))
	fmt.Fprintf(w, "  Balance  $%.2f\n", st.BalanceUSD)
	fmt.Fprintf(w, "  API      %s\n", dashIfEmpty(st.APIEndpoint))
	if len(st.Teams) > 0 {
		fmt.Fprintf(w, "\nTeams (%d)\n", len(st.Teams))
		for _, t := range st.Teams {
			marker := " "
			if t.ID == st.DefaultTeamID {
				marker = "*"
			}
			fmt.Fprintf(w, "  %s %s  %s  %s\n", marker, t.ID, t.Name, t.Role)
		}
	}
}

func newConfigureCmd() *cobra.Command {
	var endpoint, output, issuer, clientID, audienceScope string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Override the hosted-service defaults (only needed for dev/staging or self-hosted)",
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
