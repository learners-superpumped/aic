package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage AIC API keys (scoped machine credentials)",
	}
	cmd.AddCommand(newKeysCreateCmd(), newKeysListCmd(), newKeysRevokeCmd())
	return cmd
}

const teamFullAccessBanner = `WARNING: this is a TEAM FULL-ACCESS key. It can act with the team owner's
authority on every resource in this team, INCLUDING SPEND (ads, billing,
domain purchases). Store it like a password and revoke it immediately if
it is ever exposed.`

const projectFullAccessBanner = `WARNING: this is a PROJECT FULL-ACCESS key. It can act with full authority
on every primitive in the bound project, INCLUDING SPEND (ads, domain
purchases). It cannot touch other projects or team-level billing. Store it
like a password and revoke it immediately if it is ever exposed.`

func newKeysCreateCmd() *cobra.Command {
	var (
		scopes    []string
		name      string
		expiresIn string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an AIC API key (shown once — copy it immediately)",
		Long: `Create an AIC API key for server-to-server calls.

Keys carry one or more --scope values. Project-level scopes require --project.
Use --scope '*' for a full-access key: with --project it is project-owner over
that project's primitives; without --project it is team-owner over the whole
team. The full-access scope cannot be combined with other scopes.

The raw key is printed once and cannot be recovered; rotate by creating a
new key and revoking the old one.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if len(scopes) == 0 {
				return fmt.Errorf("pass at least one --scope (use --scope '*' for a full-access key)")
			}
			if isStarScope(scopes) && len(scopes) > 1 {
				return fmt.Errorf("--scope '*' cannot be combined with other scopes")
			}
			expSecs, err := parseExpiresIn(expiresIn)
			if err != nil {
				return err
			}
			// only bind if --project was explicitly passed, never from the profile default
			projectID := ""
			if cmd.Flags().Changed("project") {
				projectID = a.Project
			}
			k, err := a.Client.CreateAPIKey(cmd.Context(), a.Team, api.CreateAPIKeyRequest{
				Scopes:    scopes,
				ProjectID: projectID,
				Name:      name,
				ExpiresIn: expSecs,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if isStarScope(scopes) {
				if projectID != "" {
					fmt.Fprintln(out, projectFullAccessBanner)
				} else {
					fmt.Fprintln(out, teamFullAccessBanner)
				}
				fmt.Fprintln(out)
			}
			fmt.Fprintf(out, "Created AIC API key %s (prefix %s)\n\n  %s\n\nCopy it now — it will not be shown again.\n",
				k.ID, k.Prefix, k.Key)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&scopes, "scope", nil,
		"capability scope, repeatable (e.g. --scope storage:read --scope storage:write, or --scope '*')")
	cmd.Flags().StringVar(&name, "name", "", "human label for the key")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "",
		"key lifetime, e.g. 90d, 12h, 30m (default: no expiry; manage by revoke)")
	return cmd
}

func isStarScope(scopes []string) bool {
	for _, s := range scopes {
		if s == "*" {
			return true
		}
	}
	return false
}

func newKeysListCmd() *cobra.Command {
	var limit int
	var cursor string
	cmd := &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List the team's AIC API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			projectID := ""
			if cmd.Flags().Changed("project") {
				projectID = a.Project
			}
			page, err := a.Client.ListAPIKeys(cmd.Context(), a.Team, projectID, limit, cursor)
			if err != nil {
				return err
			}
			return printPage(cmd, a.Out, page,
				[]string{"ID", "PREFIX", "SCOPES", "PROJECT", "NAME", "STATUS", "EXPIRES", "LAST USED"},
				func(v any) []string {
					k := v.(api.APIKey)
					return []string{
						k.ID, k.Prefix, strings.Join(k.Scopes, ","), k.ProjectID, k.Name,
						k.Status, fmtKeyTime(k.ExpiresAt), fmtKeyTime(k.LastUsedAt),
					}
				})
		},
	}
	addPaginationFlags(cmd, &limit, &cursor)
	return cmd
}

func newKeysRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use: "revoke <keyId>", Short: "Revoke an AIC API key immediately", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.RevokeAPIKey(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Key revoked. Requests using it are rejected immediately.")
			return nil
		},
	}
}

// parseExpiresIn converts "90d", "12h", "30m" into seconds ("" means no expiry).
func parseExpiresIn(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid --expires-in %q: use forms like 90d, 12h, 30m", s)
		}
		return int64(n) * 86400, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid --expires-in %q: use forms like 90d, 12h, 30m", s)
	}
	return int64(d.Seconds()), nil
}

func fmtKeyTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.UTC().Format("2006-01-02")
}
