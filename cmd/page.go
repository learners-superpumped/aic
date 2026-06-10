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
// single-quoted, which suppresses every shell expansion, with each embedded
// single quote escaped via the POSIX close-escape-reopen idiom. Most list
// filters are bare (addresses, ids, enums);
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
// Commands that take positional args (e.g. `storage ls <bucket>`) pass them so
// the footer command stays runnable.
func printPage[T any](cmd *cobra.Command, out *output.Renderer, page api.Page[T], headers []string, row output.RowFunc, args ...string) error {
	if out.Format() != "table" {
		return out.Print(page, headers, row)
	}
	if err := out.Print(page.Data, headers, row); err != nil {
		return err
	}
	if page.HasMore {
		// Reconstruct a copy-pasteable next-page command that preserves every
		// flag the user set on the command line — local filters (--direction,
		// --inbox, --from, --limit, …) AND scope overrides (--team, --project,
		// --profile), since dropping either makes page 2 query a DIFFERENT
		// result set. Only --cursor is swapped for the fresh value and --output
		// is skipped (presentation-only; the footer only prints in table mode).
		// Flags are emitted as --name=value so bool flags round-trip, and slice
		// flags repeat --name=v per element (Value.String()'s "[a,b]" doesn't
		// re-parse). Visit only sees Changed flags, so config-sourced defaults
		// never leak in.
		next := cmd.CommandPath()
		for _, a := range args {
			next += " " + shellQuote(a)
		}
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Name == "cursor" || f.Name == "output" {
				return
			}
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				for _, v := range sv.GetSlice() {
					next += " --" + f.Name + "=" + shellQuote(v)
				}
				return
			}
			next += " --" + f.Name + "=" + shellQuote(f.Value.String())
		})
		fmt.Fprintf(out.Writer(), "\nMore results available. Next page:\n  %s --cursor %s\n", next, shellQuote(page.NextCursor))
	}
	return nil
}
