package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/learners-superpumped/aicompany-platform/cli/internal/auth"
	"github.com/spf13/cobra"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage payment methods and billing",
	}
	cmd.AddCommand(
		newBillingAddCardCmd(),
		newBillingCardsCmd(),
		newBillingStatusCmd(),
	)
	return cmd
}

func newBillingAddCardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-card",
		Short: "Register a card via your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			_, err = auth.RunFlow(cmd.Context(), auth.Flow{
				Start: func(ctx context.Context) (string, string, error) {
					s, err := a.Client.StartCardSession(ctx)
					if err != nil {
						return "", "", err
					}
					return s.SessionID, s.BrowserURL, nil
				},
				OpenBrowser: auth.OpenBrowser,
				Poll: func(ctx context.Context, id string) (string, error) {
					s, err := a.Client.PollCardSession(ctx, id)
					if err != nil {
						return "", err
					}
					return s.Status, nil
				},
				Interval: 2 * time.Second,
				Timeout:  5 * time.Minute,
			})
			if err != nil {
				return err
			}
			fmt.Println("Card registered successfully.")
			return nil
		},
	}
}

func newBillingCardsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cards",
		Short: "List registered cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			items, err := a.Client.ListCards(cmd.Context())
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "BRAND", "LAST4", "EXPIRES", "DEFAULT"}, func(v any) []string {
				c := v.(api.Card)
				return []string{c.CardID, c.Brand, c.Last4, fmt.Sprintf("%02d/%d", c.ExpMonth, c.ExpYear), fmt.Sprintf("%t", c.Default)}
			})
		},
	}
}

func newBillingStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show billing status",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			s, err := a.Client.BillingStatus(cmd.Context())
			if err != nil {
				return err
			}
			return a.Out.Print(*s, []string{"HAS PAYMENT METHOD"}, func(v any) []string {
				x := v.(api.BillingStatus)
				return []string{fmt.Sprintf("%t", x.HasPaymentMethod)}
			})
		},
	}
}
