package cmd

import (
	"fmt"
	"strings"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// addPaginationFlags registers the standard --limit/--cursor flags on a list
// command. One place so every list surface stays consistent.
func addPaginationFlags(cmd *cobra.Command, limit *int, cursor *string) {
	cmd.Flags().IntVar(limit, "limit", 0, "max rows per page (default 50, max 200)")
	cmd.Flags().StringVar(cursor, "cursor", "", "next page's cursor — the 'next_cursor' value from a previous list (shown in -o json output and the table footer)")
}

// shellQuote POSIX-quotes a flag value for the copy-pasteable next-page command
// so it survives a shell paste verbatim. Values made only of shell-safe chars
// pass through bare; anything else (whitespace, $, backtick, quotes, …) is
// single-quoted, which suppresses every shell expansion, with embedded single
// quotes escaped as '\''. Most list filters are bare (addresses, ids, enums);
// this only kicks in when they aren't. Note: the footer echoes the user's own
// already-run input, so this is paste fidelity, not a trust boundary.
func shellQuote(v string) string {
	const safe = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_@%+=:,./-"
	if v == "" {
		return "''"
	}
	for _, r := range v {
		if !strings.ContainsRune(safe, r) {
			return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
		}
	}
	return v
}

// printPage renders a paginated envelope. In table mode it prints the rows and,
// when another page exists, a copy-pasteable next-page command. In json/yaml
// mode it marshals the whole envelope so scripts/agents read next_cursor.
func printPage[T any](cmd *cobra.Command, out *output.Renderer, page api.Page[T], headers []string, row output.RowFunc) error {
	if out.Format() != "table" {
		return out.Print(page, headers, row)
	}
	if err := out.Print(page.Data, headers, row); err != nil {
		return err
	}
	if page.HasMore {
		// Reconstruct a copy-pasteable next-page command that preserves every
		// filter the user set (--direction, --inbox, --from, --limit, …) so the
		// next page queries the SAME result set. Only --cursor is swapped for the
		// fresh value. Visit the command's own flag set (where Changed is tracked)
		// and skip inherited globals (--output, --project, …) so only this list's
		// own flags carry forward.
		next := cmd.CommandPath()
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Name == "cursor" || cmd.InheritedFlags().Lookup(f.Name) != nil {
				return
			}
			next += " --" + f.Name + " " + shellQuote(f.Value.String())
		})
		fmt.Fprintf(out.Writer(), "\nMore results available. Next page:\n  %s --cursor %s\n", next, page.NextCursor)
	}
	return nil
}
