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

const fullAccessBanner = `WARNING: this is a FULL-ACCESS key. It can act with the team owner's
authority on every resource in this team, INCLUDING SPEND (ads, billing,
domain purchases). Store it like a password and revoke it immediately if
it is ever exposed.`

func newKeysCreateCmd() *cobra.Command {
	var (
		scopes     []string
		fullAccess bool
		name       string
		expiresIn  string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an AIC API key (shown once — copy it immediately)",
		Long: `Create an AIC API key for server-to-server calls.

Scoped keys carry one or more --scope values (all project-level or all
team-level). Project-level scopes require --project. --full-access mints a
key with the team owner's full authority instead (no --scope/--project).

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
			if fullAccess && len(scopes) > 0 {
				return fmt.Errorf("--full-access cannot be combined with --scope")
			}
			if !fullAccess && len(scopes) == 0 {
				return fmt.Errorf("pass at least one --scope, or --full-access")
			}
			expSecs, err := parseExpiresIn(expiresIn)
			if err != nil {
				return err
			}
			// Bind to a project only when --project was explicitly set on the
			// command line: a profile's default project must never silently
			// bind a credential.
			projectID := ""
			if cmd.Flags().Changed("project") {
				projectID = a.Project
			}
			if fullAccess && projectID != "" {
				return fmt.Errorf("--full-access keys are team-level; do not pass --project")
			}
			k, err := a.Client.CreateAPIKey(cmd.Context(), a.Team, api.CreateAPIKeyRequest{
				Scopes:     scopes,
				FullAccess: fullAccess,
				ProjectID:  projectID,
				Name:       name,
				ExpiresIn:  expSecs,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if fullAccess {
				fmt.Fprintln(out, fullAccessBanner)
				fmt.Fprintln(out)
			}
			fmt.Fprintf(out, "Created AIC API key %s (prefix %s)\n\n  %s\n\nCopy it now — it will not be shown again.\n",
				k.ID, k.Prefix, k.Key)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&scopes, "scope", nil,
		"capability scope, repeatable (e.g. --scope storage:read --scope storage:write)")
	cmd.Flags().BoolVar(&fullAccess, "full-access", false,
		"mint a full-access key acting with the team owner's authority (use with care)")
	cmd.Flags().StringVar(&name, "name", "", "human label for the key")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "",
		"key lifetime, e.g. 90d, 12h, 30m (default: no expiry; manage by revoke)")
	return cmd
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
