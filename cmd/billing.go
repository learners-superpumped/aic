package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/auth"
	"github.com/spf13/cobra"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage credits and payment methods",
	}
	cmd.AddCommand(
		newBillingAddCardCmd(),
		newBillingCardsCmd(),
		newBillingTopupCmd(),
		newBillingBalanceCmd(),
		newBillingHistoryCmd(),
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
			if err := a.RequireTeam(); err != nil {
				return err
			}
			_, err = auth.RunFlow(cmd.Context(), auth.Flow{
				Start: func(ctx context.Context) (string, string, error) {
					s, err := a.Client.StartCardSession(ctx, a.Team)
					if err != nil {
						return "", "", err
					}
					return s.SessionID, s.BrowserURL, nil
				},
				OpenBrowser: auth.OpenBrowser,
				Poll: func(ctx context.Context, id string) (string, error) {
					s, err := a.Client.PollCardSession(ctx, a.Team, id)
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
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.ListCards(cmd.Context(), a.Team)
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

func newBillingTopupCmd() *cobra.Command {
	var amount string
	cmd := &cobra.Command{
		Use:   "topup",
		Short: "Buy credits by charging your saved card",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			cents, err := parseDollarsToCents(amount)
			if err != nil {
				return err
			}
			res, err := a.Client.Topup(cmd.Context(), a.Team, cents)
			if err != nil {
				return err
			}
			if res.Status == "requires_action" {
				return fmt.Errorf("card requires re-authentication: run `aic billing add-card` again")
			}
			return a.Out.Print(*res, []string{"STATUS", "PAYMENT_INTENT"}, func(v any) []string {
				x := v.(api.TopupResult)
				return []string{x.Status, x.PaymentIntentID}
			})
		},
	}
	cmd.Flags().StringVar(&amount, "amount", "", "amount in USD to add, e.g. 50 or 49.99 (required)")
	_ = cmd.MarkFlagRequired("amount")
	return cmd
}

func newBillingBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Show your credit balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			b, err := a.Client.Balance(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(*b, []string{"BALANCE (USD)"}, func(v any) []string {
				x := v.(api.CreditBalance)
				return []string{fmt.Sprintf("$%.2f", x.BalanceUSD)}
			})
		},
	}
}

func newBillingHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Show your credit ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.History(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "TYPE", "AMOUNT (USD)", "REFERENCE", "WHEN"}, func(v any) []string {
				e := v.(api.LedgerEntry)
				return []string{e.ID, e.Type, fmt.Sprintf("$%.4f", float64(e.AmountNano)/1e9), e.Reference, e.CreatedAt.Format(time.RFC3339)}
			})
		},
	}
}

// parseDollarsToCents converts a dollar string ("50", "49.99") to integer cents
// without floating-point rounding error.
func parseDollarsToCents(s string) (int64, error) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "$"))
	if s == "" {
		return 0, fmt.Errorf("amount is required")
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("amount must be positive")
	}
	parts := strings.SplitN(s, ".", 2)
	dollars, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || dollars < 0 {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	var cents int64
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			return 0, fmt.Errorf("amount %q has more than two decimal places", s)
		}
		for len(frac) < 2 {
			frac += "0"
		}
		cents, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
	}
	total := dollars*100 + cents
	if total <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	return total, nil
}
