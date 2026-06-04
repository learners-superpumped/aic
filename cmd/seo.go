package cmd

import (
	"fmt"

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
				return a.Out.Print(site, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
					s := v.(api.SEOSiteDTO)
					return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
				})
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
		&cobra.Command{
			Use: "ls", Short: "List sites", Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				sites, err := a.Client.ListSEOSites(cmd.Context(), a.Team, a.Project)
				if err != nil {
					return err
				}
				return a.Out.Print(sites, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
					s := v.(api.SEOSiteDTO)
					return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
				})
			},
		},
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
				return a.Out.Print(site, []string{"DOMAIN", "DNS", "STATUS"}, func(v any) []string {
					s := v.(api.SEOSiteDTO)
					return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
				})
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
	var dims []string
	var limit int
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
			rows, err := a.Client.SCQuery(cmd.Context(), a.Team, a.Project, args[0],
				api.SCQuery{StartDate: start, EndDate: end, Dimensions: dims, Limit: limit})
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
	queryCmd.Flags().IntVar(&limit, "limit", 0, "row limit")
	cmd.AddCommand(queryCmd)

	cmd.AddCommand(&cobra.Command{
		Use: "sitemaps <domain>", Short: "List sitemaps", Args: cobra.ExactArgs(1),
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

func boolYN(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
