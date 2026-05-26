package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/learners-company/aic/internal/api"
	"github.com/learners-company/aic/internal/auth"
	"github.com/learners-company/aic/internal/config"
	"github.com/spf13/cobra"
)

func newAuthCmds() []*cobra.Command {
	return []*cobra.Command{
		newLoginCmd(),
		newLogoutCmd(),
		newWhoamiCmd(),
		newConfigureCmd(),
	}
}

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via your browser and store credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName, _ := cmd.Flags().GetString("profile")

			prof, _ := config.Load(profileName)
			endpoint := defaultEndpoint
			if prof != nil && prof.APIEndpoint != "" {
				endpoint = prof.APIEndpoint
			}
			client := api.New(endpoint, "")

			var tokens *api.Tokens
			_, err := auth.RunFlow(cmd.Context(), auth.Flow{
				Start: func(ctx context.Context) (string, string, error) {
					s, err := client.StartLoginSession(ctx)
					if err != nil {
						return "", "", err
					}
					return s.SessionID, s.BrowserURL, nil
				},
				OpenBrowser: auth.OpenBrowser,
				Poll: func(ctx context.Context, id string) (string, error) {
					s, err := client.PollLoginSession(ctx, id)
					if err != nil {
						return "", err
					}
					if s.Status == "completed" {
						tokens = s.Tokens
					}
					return s.Status, nil
				},
				Interval: 2 * time.Second,
				Timeout:  5 * time.Minute,
			})
			if err != nil {
				return err
			}
			if tokens == nil {
				return fmt.Errorf("login completed but no tokens were returned")
			}

			save := &config.Profile{
				Name:         profileName,
				AccessToken:  tokens.AccessToken,
				RefreshToken: tokens.RefreshToken,
				ExpiresAt:    tokens.ExpiresAt,
				APIEndpoint:  endpoint,
				Output:       "table",
			}
			if prof != nil {
				save.DefaultProject = prof.DefaultProject
				if prof.Output != "" {
					save.Output = prof.Output
				}
			}
			if err := config.Save(save); err != nil {
				return err
			}
			fmt.Println("Login successful. Credentials saved.")
			return nil
		},
	}
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
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			id, err := a.Client.Whoami(cmd.Context())
			if err != nil {
				return err
			}
			return a.Out.Print(*id, []string{"USER ID", "EMAIL"}, func(v any) []string {
				x := v.(api.Identity)
				return []string{x.UserID, x.Email}
			})
		},
	}
}

func newConfigureCmd() *cobra.Command {
	var endpoint, output string
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
			if err := config.Save(prof); err != nil {
				return err
			}
			fmt.Println("Configuration saved.")
			return nil
		},
	}
	cmd.Flags().StringVar(&endpoint, "api-endpoint", "", "backend API endpoint URL")
	cmd.Flags().StringVar(&output, "output-format", "", "default output format: table|json|yaml")
	return cmd
}
