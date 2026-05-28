package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
	"github.com/spf13/cobra"
)

func newMailDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "domains", Short: "Manage SES domain identities"}
	cmd.AddCommand(
		newMailDomainsEnableCmd(),
		newMailDomainsShowCmd(),
		newMailDomainsVerifyCmd(),
		newMailDomainsListCmd(),
		newMailDomainsDisableCmd(),
	)
	return cmd
}

func newMailDomainsEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <domain>",
		Short: "Enable a domain for outbound mail (SES identity + DKIM)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			out, err := a.Client.EnableMailDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			return printDomainResponse(a, out)
		},
	}
}

func printDomainResponse(a *app.App, out *api.EnableMailDomainResponse) error {
	fmt.Printf("Domain:        %s\nStatus:        %s\nMAIL FROM:     %s\nAuto-applied:  %v\n",
		out.Identity.Name, out.Identity.Status, out.Identity.MailFromDomain, out.AutoApplied)
	if !out.AutoApplied {
		fmt.Println("\nAdd these DNS records at your provider:")
	} else {
		fmt.Println("\nWe applied these records to your Route 53 hosted zone:")
	}
	return a.Out.Print(out.Records, []string{"NAME", "TYPE", "VALUE", "TTL"}, func(v any) []string {
		r := v.(api.MailDNSRecord)
		return []string{r.Name, r.Type, r.Value, fmt.Sprintf("%d", r.TTL)}
	})
}

func newMailDomainsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <domain>",
		Short: "Show identity status and DNS records",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			out, err := a.Client.ShowMailDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			return printDomainResponse(a, out)
		},
	}
}

func newMailDomainsVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <domain>",
		Short: "Force an immediate verification re-check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			out, err := a.Client.VerifyMailDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			return printDomainResponse(a, out)
		},
	}
}

func newMailDomainsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List email-enabled domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			items, err := a.Client.ListMailDomains(cmd.Context(), a.Team, a.Project)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"DOMAIN", "STATUS", "VERIFIED_AT"}, func(v any) []string {
				e := v.(api.MailIdentity)
				return []string{e.Name, e.Status, e.VerifiedAt}
			})
		},
	}
}

func newMailDomainsDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <domain>",
		Short: "Disable a domain (deletes identity + DKIM records)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if err := a.Client.DisableMailDomain(cmd.Context(), a.Team, a.Project, args[0]); err != nil {
				return err
			}
			fmt.Printf("Mail disabled for %s\n", args[0])
			return nil
		},
	}
}
