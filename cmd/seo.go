package cmd

import (
	"fmt"
	"strings"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newSEOCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "seo", Short: "Manage AIC SEO integrations (Search Console)"}
	cmd.AddCommand(newSEOSitesCmd(), newSearchConsoleCmd())
	return cmd
}

func newSEOSitesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "sites", Short: "Manage SEO sites"}
	var lsLimit int
	var lsCursor string
	lsCmd := &cobra.Command{
		Use: "ls", Short: "List sites", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			page, err := a.Client.ListSEOSites(cmd.Context(), a.Team, a.Project, lsLimit, lsCursor)
			if err != nil {
				return err
			}
			return printPage(cmd, a.Out, page, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
				s := v.(api.SEOSiteDTO)
				return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
			})
		},
	}
	addPaginationFlags(lsCmd, &lsLimit, &lsCursor)
	cmd.AddCommand(
		&cobra.Command{
			Use: "add <domain>", Short: "Register a site", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				site, err := a.Client.AddSEOSite(cmd.Context(), a.Team, a.Project, args[0])
				if err != nil {
					return err
				}
				if err := a.Out.Print(site, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
					s := v.(api.SEOSiteDTO)
					return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
				}); err != nil {
					return err
				}
				printVerifyInstruction(cmd, site)
				return nil
			},
		},
		&cobra.Command{
			Use: "verify <domain>", Short: "Verify ownership (external domains)", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				status, err := a.Client.VerifySEOSite(cmd.Context(), a.Team, a.Project, args[0])
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), status)
				return nil
			},
		},
		lsCmd,
		&cobra.Command{
			Use: "show <domain>", Short: "Show a site", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				site, err := a.Client.ShowSEOSite(cmd.Context(), a.Team, a.Project, args[0])
				if err != nil {
					return err
				}
				if err := a.Out.Print(site, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
					s := v.(api.SEOSiteDTO)
					return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
				}); err != nil {
					return err
				}
				printVerifyInstruction(cmd, site)
				return nil
			},
		},
		&cobra.Command{
			Use: "rm <domain>", Short: "Deregister a site", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				return a.Client.DeleteSEOSite(cmd.Context(), a.Team, a.Project, args[0])
			},
		},
	)
	return cmd
}

func newSearchConsoleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "search-console", Short: "Query Search Console data"}
	var start, end string
	var dims, filters []string
	var searchType string
	var limit, startRow int
	queryCmd := &cobra.Command{
		Use: "query <domain>", Short: "Search analytics", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			var fs []api.SCFilter
			for _, raw := range filters {
				f, perr := parseFilter(raw)
				if perr != nil {
					return perr
				}
				fs = append(fs, f)
			}
			rows, err := a.Client.SCQuery(cmd.Context(), a.Team, a.Project, args[0],
				api.SCQuery{StartDate: start, EndDate: end, Dimensions: dims, Filters: fs,
					Type: searchType, StartRow: startRow, Limit: limit})
			if err != nil {
				return err
			}
			return a.Out.Print(rows, []string{"DATE", "CLICKS", "IMPRESSIONS", "CTR", "POSITION"}, func(v any) []string {
				m := v.(api.SCMetricDTO)
				return []string{m.Date, fmt.Sprint(m.Clicks), fmt.Sprint(m.Impressions), fmt.Sprintf("%.3f", m.CTR), fmt.Sprintf("%.1f", m.Position)}
			})
		},
	}
	queryCmd.Flags().StringVar(&start, "start", "", "start date YYYY-MM-DD")
	queryCmd.Flags().StringVar(&end, "end", "", "end date YYYY-MM-DD")
	queryCmd.Flags().StringSliceVar(&dims, "dimensions", nil, "query|page|country|device")
	queryCmd.Flags().StringArrayVar(&filters, "filter", nil, `dimension filter "<dim> <op> <expr>", e.g. "country equals usa" (repeatable)`)
	queryCmd.Flags().StringVar(&searchType, "type", "", "web|image|video|news|discover|googleNews (default web)")
	queryCmd.Flags().IntVar(&startRow, "start-row", 0, "pagination offset (0 = first page)")
	queryCmd.Flags().IntVar(&limit, "limit", 0, "row limit")
	cmd.AddCommand(queryCmd)

	sitemaps := &cobra.Command{Use: "sitemaps", Short: "List, submit, or delete sitemaps"}
	sitemaps.AddCommand(&cobra.Command{
		Use: "list <domain>", Short: "List submitted sitemaps", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			sm, err := a.Client.SCSitemaps(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(sm, []string{"PATH", "SUBMITTED", "ERRORS", "WARNINGS"}, func(v any) []string {
				s := v.(api.SCSitemapDTO)
				return []string{s.Path, s.LastSubmitted, fmt.Sprint(s.Errors), fmt.Sprint(s.Warnings)}
			})
		},
	})
	sitemaps.AddCommand(&cobra.Command{
		Use: "submit <domain> <sitemap-url>", Short: "Submit a sitemap", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if err := a.Client.SCSubmitSitemap(cmd.Context(), a.Team, a.Project, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "submitted %s\n", args[1])
			return nil
		},
	})
	sitemaps.AddCommand(&cobra.Command{
		Use: "delete <domain> <sitemap-url>", Short: "Delete a submitted sitemap", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if err := a.Client.SCDeleteSitemap(cmd.Context(), a.Team, a.Project, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", args[1])
			return nil
		},
	})
	cmd.AddCommand(sitemaps)

	cmd.AddCommand(&cobra.Command{
		Use: "inspect <domain> <url>", Short: "Inspect a URL's index status", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			ins, err := a.Client.SCInspect(cmd.Context(), a.Team, a.Project, args[0], args[1])
			if err != nil {
				return err
			}
			return a.Out.Print(ins, []string{"URL", "VERDICT", "COVERAGE", "LAST_CRAWL"}, func(v any) []string {
				i := v.(api.SCInspectionDTO)
				return []string{i.URL, i.Verdict, i.CoverageState, i.LastCrawlTime}
			})
		},
	})
	return cmd
}

func printVerifyInstruction(cmd *cobra.Command, s api.SEOSiteDTO) {
	if s.VerifyRecordValue == "" {
		return
	}
	w := cmd.ErrOrStderr()
	fmt.Fprintf(w, "\nAdd this DNS TXT record, then run: aic seo sites verify %s\n", s.Domain)
	fmt.Fprintf(w, "  name:  %s\n", s.VerifyRecordName)
	fmt.Fprintf(w, "  value: %s\n", s.VerifyRecordValue)
}

func parseFilter(s string) (api.SCFilter, error) {
	parts := strings.SplitN(strings.TrimSpace(s), " ", 3)
	if len(parts) != 3 {
		return api.SCFilter{}, fmt.Errorf("filter must be \"<dimension> <operator> <expression>\", got %q", s)
	}
	return api.SCFilter{Dimension: parts[0], Operator: parts[1], Expression: parts[2]}, nil
}

func boolYN(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
