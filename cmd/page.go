package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/output"
	"github.com/spf13/cobra"
)

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
		next := cmd.CommandPath()
		if f := cmd.Flags().Lookup("limit"); f != nil && f.Changed {
			next += " --limit " + f.Value.String()
		}
		fmt.Fprintf(out.Writer(), "\nMore results available. Next page:\n  %s --cursor %s\n", next, page.NextCursor)
	}
	return nil
}
