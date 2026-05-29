package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newDomainsRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "records", Aliases: []string{"rec", "dns"}, Short: "Manage DNS records for a connected domain"}
	cmd.AddCommand(newRecordsListCmd(), newRecordsAddCmd(), newRecordsSetCmd(), newRecordsDeleteCmd(), newRecordsImportCmd())
	return cmd
}

func recordRows() ([]string, func(any) []string) {
	return []string{"NAME", "TYPE", "VALUES", "TTL", "SOURCE"}, func(v any) []string {
		r := v.(api.DNSRecord)
		val := ""
		for i, x := range r.Values {
			if i > 0 {
				val += ", "
			}
			val += x
		}
		return []string{r.Name, r.Type, val, fmt.Sprintf("%d", r.TTL), r.Source}
	}
}

func newRecordsListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list <domain>", Aliases: []string{"ls"}, Short: "List DNS records", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			items, err := a.Client.ListDNSRecords(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			hdr, row := recordRows()
			return a.Out.Print(items, hdr, row)
		},
	}
}

func newRecordsAddCmd() *cobra.Command {
	var rtype, name string
	var values []string
	var ttl int32
	cmd := &cobra.Command{
		Use: "add <domain>", Short: "Add a DNS record (fails if it already exists)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			r := api.DNSRecord{Type: rtype, Name: recordName(name, args[0]), Values: values, TTL: ttl}
			out, err := a.Client.AddDNSRecord(cmd.Context(), a.Team, a.Project, args[0], r)
			if err != nil {
				return err
			}
			hdr, row := recordRows()
			return a.Out.Print(*out, hdr, row)
		},
	}
	recordFlags(cmd, &rtype, &name, &values, &ttl)
	return cmd
}

func newRecordsSetCmd() *cobra.Command {
	var rtype, name string
	var values []string
	var ttl int32
	cmd := &cobra.Command{
		Use: "set <domain>", Short: "Create or replace a DNS record set (UPSERT)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			r := api.DNSRecord{Type: rtype, Name: recordName(name, args[0]), Values: values, TTL: ttl}
			out, err := a.Client.SetDNSRecord(cmd.Context(), a.Team, a.Project, args[0], r)
			if err != nil {
				return err
			}
			hdr, row := recordRows()
			return a.Out.Print(*out, hdr, row)
		},
	}
	recordFlags(cmd, &rtype, &name, &values, &ttl)
	return cmd
}

func newRecordsDeleteCmd() *cobra.Command {
	var rtype, name string
	cmd := &cobra.Command{
		Use: "delete <domain>", Aliases: []string{"rm"}, Short: "Delete a DNS record set", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			if err := a.Client.DeleteDNSRecord(cmd.Context(), a.Team, a.Project, args[0], recordName(name, args[0]), rtype); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s %s\n", recordName(name, args[0]), rtype)
			return nil
		},
	}
	cmd.Flags().StringVar(&rtype, "type", "", "record type (A, CNAME, MX, TXT, ...)")
	cmd.Flags().StringVar(&name, "name", "", "record name; @ for apex")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newRecordsImportCmd() *cobra.Command {
	return &cobra.Command{
		Use: "import <domain>", Short: "Best-effort scan existing records and apply to the zone", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := domainScope(a); err != nil {
				return err
			}
			items, warnings, err := a.Client.ImportDNSRecords(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			for _, w := range warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠ %s\n", w)
			}
			hdr, row := recordRows()
			return a.Out.Print(items, hdr, row)
		},
	}
}

func recordFlags(cmd *cobra.Command, rtype, name *string, values *[]string, ttl *int32) {
	cmd.Flags().StringVar(rtype, "type", "", "record type (A, AAAA, CNAME, MX, TXT, CAA, SRV)")
	cmd.Flags().StringVar(name, "name", "", "record name; @ for apex")
	cmd.Flags().StringArrayVar(values, "value", nil, "record value (repeatable for multi-value sets)")
	cmd.Flags().Int32Var(ttl, "ttl", 300, "TTL in seconds")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("value")
}

func recordName(name, domain string) string {
	if name == "@" || name == "" {
		return domain
	}
	if name == domain || hasDotSuffix(name, domain) {
		return name
	}
	return name + "." + domain
}

func hasDotSuffix(s, suffix string) bool {
	return len(s) > len(suffix)+1 && s[len(s)-len(suffix)-1:] == "."+suffix
}
