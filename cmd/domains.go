package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/spf13/cobra"
)

func newDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Search, buy, and manage domains",
	}
	cmd.AddCommand(
		newDomainsSearchCmd(),
		newDomainsBuyCmd(),
		newDomainsListCmd(),
		newDomainsShowCmd(),
	)
	return cmd
}

func newDomainsSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search domain availability and pricing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			results, err := a.Client.SearchDomains(cmd.Context(), a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(results, []string{"DOMAIN", "AVAILABLE", "PRICE"}, func(v any) []string {
				d := v.(api.DomainSearchResult)
				return []string{d.Domain, fmt.Sprintf("%t", d.Available), fmt.Sprintf("%.2f %s", d.Price, d.Currency)}
			})
		},
	}
}

func newDomainsBuyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "buy <domain>",
		Short: "Purchase a domain in the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			d, err := a.Client.BuyDomain(cmd.Context(), a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*d, []string{"DOMAIN", "STATUS"}, func(v any) []string {
				x := v.(api.Domain)
				return []string{x.Domain, x.Status}
			})
		},
	}
}

func newDomainsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List domains in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			items, err := a.Client.ListDomains(cmd.Context(), a.Project)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"DOMAIN", "STATUS"}, func(v any) []string {
				x := v.(api.Domain)
				return []string{x.Domain, x.Status}
			})
		},
	}
}

func newDomainsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <domain>",
		Short: "Show a domain in the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			d, err := a.Client.GetDomain(cmd.Context(), a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*d, []string{"DOMAIN", "STATUS"}, func(v any) []string {
				x := v.(api.Domain)
				return []string{x.Domain, x.Status}
			})
		},
	}
}
