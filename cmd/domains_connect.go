package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDomainsConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect <domain>",
		Short: "Bring an externally-registered domain under our DNS (creates a Route 53 hosted zone)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			res, err := a.Client.ConnectDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			if a.Out.Format() != "table" {
				return a.Out.Print(*res, nil, nil)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Connected %s (status=%s)\n", res.Domain.Name, res.Domain.Status)
			fmt.Fprintln(cmd.OutOrStdout(), "Set these 4 NS records at your registrar:")
			for _, ns := range res.Domain.Nameservers {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", ns)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "We will detect the change automatically.")
			fmt.Fprintf(cmd.OutOrStdout(), "Run 'aic domains verify %s' to recheck now.\n", res.Domain.Name)
			return nil
		},
	}
}

func newDomainsVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <domain>",
		Short: "Check NS propagation for a connected domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			res, err := a.Client.VerifyDomain(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			if a.Out.Format() != "table" {
				return a.Out.Print(*res, nil, nil)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s (checked at %s)\n", res.Domain.Status, res.CheckedAt.Format("2006-01-02 15:04:05 UTC"))
			if res.Domain.Status != "verified" {
				fmt.Fprintln(cmd.OutOrStdout(), "Expected:")
				for _, ns := range res.Expected {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", ns)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Observed:")
				for _, ns := range res.Observed {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", ns)
				}
			}
			return nil
		},
	}
}

func newDomainsDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect <domain>",
		Short: "Remove a connected domain and delete its Route 53 hosted zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			if err := a.Client.DisconnectDomain(cmd.Context(), a.Team, a.Project, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected %s\n", args[0])
			return nil
		},
	}
}
