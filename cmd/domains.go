package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
	"github.com/spf13/cobra"
)

func newDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "domains", Aliases: []string{"domain"}, Short: "Search, buy, and renew domains"}
	cmd.AddCommand(
		newDomainsSearchCmd(),
		newDomainsBuyCmd(),
		newDomainsConnectCmd(),
		newDomainsVerifyCmd(),
		newDomainsDisconnectCmd(),
		newDomainsRenewCmd(),
		newDomainsListCmd(),
		newDomainsShowCmd(),
		newDomainsContactCmd(),
	)
	return cmd
}

func domainScope(a *app.App) error {
	if err := a.RequireTeam(); err != nil {
		return err
	}
	return a.RequireProject()
}

func domainRows() ([]string, func(any) []string) {
	return []string{"NAME", "SOURCE", "STATUS", "AUTO-RENEW", "EXPIRES"}, func(v any) []string {
		d := v.(api.Domain)
		exp := ""
		if !d.ExpiresAt.IsZero() {
			exp = d.ExpiresAt.Format("2006-01-02")
		}
		return []string{d.Name, d.Source, d.Status, fmt.Sprintf("%t", d.AutoRenew), exp}
	}
}

func newDomainsSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use: "search <query>", Short: "Search availability and pricing", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			res, err := a.Client.SearchDomains(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(res, []string{"DOMAIN", "AVAILABLE", "PRICE (USD)"}, func(v any) []string {
				d := v.(api.DomainSearchResult)
				return []string{d.Domain, fmt.Sprintf("%t", d.Available), fmt.Sprintf("$%.2f", d.PriceUSD)}
			})
		},
	}
}

func newDomainsBuyCmd() *cobra.Command {
	var years int
	var autoRenew bool
	var contactName string
	cmd := &cobra.Command{
		Use: "buy <domain>", Short: "Buy a domain (charges team credits)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			d, err := a.Client.BuyDomain(cmd.Context(), a.Team, a.Project, args[0], years, autoRenew, contactName)
			if err != nil {
				return err
			}
			cols, row := domainRows()
			return a.Out.Print(*d, cols, row)
		},
	}
	cmd.Flags().IntVar(&years, "years", 1, "registration years (1-10)")
	cmd.Flags().BoolVar(&autoRenew, "auto-renew", false, "store an auto-renew preference (automatic renewal ships in a later release; use `aic domains renew` meanwhile)")
	cmd.Flags().StringVar(&contactName, "contact", "", "WHOIS contact profile to use (defaults to the team's default; manage with `aic domains contact`)")
	return cmd
}

func newDomainsRenewCmd() *cobra.Command {
	var years int
	cmd := &cobra.Command{
		Use: "renew <domain>", Short: "Renew a domain (charges team credits)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			d, err := a.Client.RenewDomain(cmd.Context(), a.Team, a.Project, args[0], years)
			if err != nil {
				return err
			}
			cols, row := domainRows()
			return a.Out.Print(*d, cols, row)
		},
	}
	cmd.Flags().IntVar(&years, "years", 1, "renewal years (1-10)")
	return cmd
}

func newDomainsListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List domains in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			items, err := a.Client.ListDomains(cmd.Context(), a.Team, a.Project)
			if err != nil {
				return err
			}
			cols, row := domainRows()
			return a.Out.Print(items, cols, row)
		},
	}
}

func newDomainsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use: "show <domain>", Short: "Show a domain", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			d, err := a.Client.GetDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			cols, row := domainRows()
			if err := a.Out.Print(*d, cols, row); err != nil {
				return err
			}
			if d.Source == "connected" {
				if len(d.Nameservers) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Nameservers:")
					for _, ns := range d.Nameservers {
						fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", ns)
					}
				}
				if !d.LastVerifyAt.IsZero() {
					fmt.Fprintf(cmd.OutOrStdout(), "Last verify: %s\n", d.LastVerifyAt.Format("2006-01-02 15:04:05 UTC"))
				}
				if !d.VerifiedAt.IsZero() {
					fmt.Fprintf(cmd.OutOrStdout(), "Verified at: %s\n", d.VerifiedAt.Format("2006-01-02 15:04:05 UTC"))
				}
			}
			return nil
		},
	}
}
