package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func adCampaignRows() ([]string, func(any) []string) {
	return []string{"ID", "STATUS", "OBJECTIVE", "BUDGET-TYPE", "BUDGET-NANO", "PROVIDER", "CREATED"},
		func(v any) []string {
			c := v.(api.AdCampaign)
			return []string{c.ID, c.Status, c.Objective, c.BudgetType, strconv.FormatInt(c.BudgetNano, 10), c.Provider, c.CreatedAt.Format(time.RFC3339)}
		}
}

func parseAgeRange(s string) (min, max int, err error) {
	if s == "" {
		return 0, 0, fmt.Errorf("age range must be in the form MIN-MAX, e.g. 25-44")
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, fmt.Errorf("age range must be in the form MIN-MAX, e.g. 25-44")
	}
	minVal, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("age range min %q is not a number", parts[0])
	}
	maxVal, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("age range max %q is not a number", parts[1])
	}
	return minVal, maxVal, nil
}

func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("ads: rand.Read: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func newAdsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "ads", Short: "Manage AIC ad campaigns"}
	cmd.AddCommand(
		newAdsLaunchCmd(),
		newAdsListCmd(),
		newAdsStatusCmd(),
		newAdsPauseCmd(),
		newAdsResumeCmd(),
		newAdsDeleteCmd(),
	)
	return cmd
}

func newAdsLaunchCmd() *cobra.Command {
	var (
		objective       string
		budgetType      string
		budgetNano      int64
		geo             []string
		age             string
		interests       []string
		genders         []string
		creativeAsset   string
		headline        string
		body            string
		cta             string
		destinationURL  string
		provider        string
		providerOptions string
		endAt           string
		launchToken     string
		placements      []string
	)

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch an ad campaign",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}

			tok := launchToken
			if tok == "" {
				tok = randomToken()
			}

			req := api.AdLaunchRequest{
				LaunchToken: tok,
				Objective:   objective,
				BudgetType:  budgetType,
				BudgetNano:  budgetNano,
				StartAt:     time.Now().UTC(),
				Placements:  placements,
				Creative: api.AdCreative{
					StorageRef:     creativeAsset,
					Headline:       headline,
					Body:           body,
					CTA:            cta,
					DestinationURL: destinationURL,
				},
			}

			if endAt != "" {
				t, err := time.Parse(time.RFC3339, endAt)
				if err != nil {
					return fmt.Errorf("--end must be RFC3339, e.g. 2026-12-31T00:00:00Z: %w", err)
				}
				req.EndAt = &t
			} else if budgetType == "lifetime" {
				return fmt.Errorf("--end is required when --budget-type is lifetime")
			}

			targeting := api.AdTargeting{
				Geo:       geo,
				Interests: interests,
				Genders:   genders,
			}
			if age != "" {
				mn, mx, err := parseAgeRange(age)
				if err != nil {
					return fmt.Errorf("--age: %w", err)
				}
				targeting.AgeMin = mn
				targeting.AgeMax = mx
			}
			req.Targeting = targeting

			if providerOptions != "" {
				var opts map[string]any
				if err := json.Unmarshal([]byte(providerOptions), &opts); err != nil {
					return fmt.Errorf("--provider-options: %w", err)
				}
				req.ProviderOptions = opts
			}

			_ = provider

			c, err := a.Client.LaunchAd(cmd.Context(), a.Team, a.Project, req)
			if err != nil {
				return err
			}
			cols, row := adCampaignRows()
			return a.Out.Print(c, cols, row)
		},
	}

	cmd.Flags().StringVar(&objective, "objective", "traffic", "campaign objective: traffic|conversions|awareness|engagement|leads")
	cmd.Flags().StringVar(&budgetType, "budget-type", "daily", "budget type: daily|lifetime")
	cmd.Flags().Int64Var(&budgetNano, "budget", 0, "budget in nano-dollars (1 USD = 1 000 000 000)")
	cmd.Flags().StringArrayVar(&geo, "geo", nil, "target country/region codes (repeatable)")
	cmd.Flags().StringVar(&age, "age", "", "target age range, e.g. 25-44")
	cmd.Flags().StringArrayVar(&interests, "interests", nil, "comma-separated interest keywords")
	cmd.Flags().StringArrayVar(&genders, "genders", nil, "target genders (repeatable)")
	cmd.Flags().StringVar(&creativeAsset, "creative-asset", "", "storage reference for the creative asset (bucket/key)")
	cmd.Flags().StringVar(&headline, "headline", "", "ad headline")
	cmd.Flags().StringVar(&body, "body", "", "ad body copy")
	cmd.Flags().StringVar(&cta, "cta", "", "call-to-action label, e.g. 'Learn More'")
	cmd.Flags().StringVar(&destinationURL, "url", "", "destination URL for the ad")
	cmd.Flags().StringVar(&provider, "provider", "meta", "ad provider")
	cmd.Flags().StringVar(&providerOptions, "provider-options", "", "provider-specific options as a JSON object")
	cmd.Flags().StringVar(&endAt, "end", "", "campaign end time (RFC3339); required for lifetime budgets")
	cmd.Flags().StringVar(&launchToken, "launch-token", "", "idempotency token (auto-generated if omitted)")
	cmd.Flags().StringArrayVar(&placements, "placements", nil, "ad placements (repeatable)")

	return cmd
}

func newAdsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List ad campaigns",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			cs, err := a.Client.ListAds(cmd.Context(), a.Team, a.Project)
			if err != nil {
				return err
			}
			cols, row := adCampaignRows()
			return a.Out.Print(cs, cols, row)
		},
	}
}

func newAdsStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <campaign-id>",
		Short: "Show status of an ad campaign",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			c, err := a.Client.GetAd(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			cols, row := adCampaignRows()
			return a.Out.Print(c, cols, row)
		},
	}
}

func newAdsPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <campaign-id>",
		Short: "Pause a running ad campaign",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			return a.Client.PauseAd(cmd.Context(), a.Team, a.Project, args[0])
		},
	}
}

func newAdsResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <campaign-id>",
		Short: "Resume a paused ad campaign",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			return a.Client.ResumeAd(cmd.Context(), a.Team, a.Project, args[0])
		},
	}
}

func newAdsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <campaign-id>",
		Short: "Delete an ad campaign",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			return a.Client.DeleteAd(cmd.Context(), a.Team, a.Project, args[0])
		},
	}
}
