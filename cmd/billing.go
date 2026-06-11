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
		Short: "Credits, payment methods, and auto-recharge",
	}
	cmd.AddCommand(
		newBillingAddCardCmd(),
		newBillingCardsCmd(),
		newBillingTopupCmd(),
		newBillingBalanceCmd(),
		newBillingHistoryCmd(),
		newBillingUsageCmd(),
		newBillingAutoRechargeCmd(),
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
			return printAction(a, actionResult{Status: "registered"}, "Card registered successfully.")
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
	var limit int
	var cursor string
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your AIC credit ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			page, err := a.Client.History(cmd.Context(), a.Team, limit, cursor)
			if err != nil {
				return err
			}
			cols := []string{"ID", "TYPE", "RESOURCE", "AMOUNT (USD)", "REFERENCE", "WHEN"}
			return printPage(cmd, a.Out, page, cols, func(v any) []string {
				e := v.(api.LedgerEntry)
				resource := e.ResourceType
				if resource == "" {
					resource = "—"
				}
				return []string{e.ID, e.Type, resource, fmt.Sprintf("$%.4f", float64(e.AmountNano)/1e9), e.Reference, e.CreatedAt.Format(time.RFC3339)}
			})
		},
	}
	addPaginationFlags(cmd, &limit, &cursor)
	return cmd
}

func newBillingUsageCmd() *cobra.Command {
	var from, to string
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show AIC credit usage by resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			summary, err := a.Client.Usage(cmd.Context(), a.Team, from, to)
			if err != nil {
				return err
			}
			if a.Out.Format() != "table" {
				return a.Out.Print(*summary, nil, nil)
			}
			cols := []string{"RESOURCE", "ENTRIES", "SPEND (USD)"}
			if err := a.Out.Print(summary.ByResource, cols, func(v any) []string {
				b := v.(api.UsageBucket)
				return []string{b.Resource, strconv.FormatInt(b.Entries, 10), b.SpendUSD}
			}); err != nil {
				return err
			}
			w := a.Out.Writer()
			fmt.Fprintln(w, strings.Repeat("─", 33))
			fmt.Fprintf(w, "%-10s %-9s %s\n", "TOTAL", "", summary.TotalSpendUSD)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start date YYYY-MM-DD (default 30 days ago)")
	cmd.Flags().StringVar(&to, "to", "", "end date YYYY-MM-DD (default today)")
	return cmd
}

func newBillingAutoRechargeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto-recharge",
		Short: "Manage automatic credit top-up",
	}
	cmd.AddCommand(
		newBillingAutoRechargeConfigCmd(),
		newBillingAutoRechargeEnableCmd(),
		newBillingAutoRechargeDisableCmd(),
	)
	return cmd
}

func newBillingAutoRechargeConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show auto-recharge configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			cfg, err := a.Client.GetAutoRecharge(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			return a.Out.Print(*cfg, []string{"ENABLED", "THRESHOLD (USD)", "AMOUNT (USD)", "MONTHLY LIMIT (USD)"}, func(v any) []string {
				c := v.(api.AutoRechargeConfig)
				return []string{
					fmt.Sprintf("%t", c.Enabled),
					fmt.Sprintf("$%.2f", c.ThresholdUSD),
					fmt.Sprintf("$%.2f", c.AmountUSD),
					fmt.Sprintf("$%.2f", c.MonthlyLimitUSD),
				}
			})
		},
	}
}

func newBillingAutoRechargeEnableCmd() *cobra.Command {
	var threshold, amount, monthlyLimit string
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable automatic credit top-up",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			threshCents, err := parseDollarsToCents(threshold)
			if err != nil {
				return fmt.Errorf("--threshold: %w", err)
			}
			amtCents, err := parseDollarsToCents(amount)
			if err != nil {
				return fmt.Errorf("--amount: %w", err)
			}
			limitCents, err := parseDollarsToCents(monthlyLimit)
			if err != nil {
				return fmt.Errorf("--monthly-limit: %w", err)
			}
			cfg := api.AutoRechargeConfig{
				Enabled:         true,
				ThresholdUSD:    float64(threshCents) / 100,
				AmountUSD:       float64(amtCents) / 100,
				MonthlyLimitUSD: float64(limitCents) / 100,
			}
			if err := a.Client.SetAutoRecharge(cmd.Context(), a.Team, cfg); err != nil {
				return err
			}
			return printAction(a, actionResult{Status: "enabled"}, "Auto-recharge enabled.")
		},
	}
	cmd.Flags().StringVar(&threshold, "threshold", "", "top up when balance drops below this amount in USD (required)")
	cmd.Flags().StringVar(&amount, "amount", "", "amount in USD to add each time (required)")
	cmd.Flags().StringVar(&monthlyLimit, "monthly-limit", "", "maximum amount in USD to auto-recharge per calendar month (required)")
	_ = cmd.MarkFlagRequired("threshold")
	_ = cmd.MarkFlagRequired("amount")
	_ = cmd.MarkFlagRequired("monthly-limit")
	return cmd
}

func newBillingAutoRechargeDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable automatic credit top-up",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.SetAutoRecharge(cmd.Context(), a.Team, api.AutoRechargeConfig{Enabled: false}); err != nil {
				return err
			}
			return printAction(a, actionResult{Status: "disabled"}, "Auto-recharge disabled.")
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
