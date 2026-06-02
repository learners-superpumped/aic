// Command docs generates the aic CLI command reference as Starlight-flavored
// Markdown. Pages are laid out in a directory tree that mirrors the command
// tree, so Starlight's `autogenerate` sidebar renders one collapsible group per
// command group (billing, domains, domains/records, ...).
//
//	go run ./tools/docs [outDir]   # default: ./reference-out
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/learners-superpumped/aic/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	outDir := "./reference-out"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.RemoveAll(outDir); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}

	root := cmd.NewRootCmd()
	root.PersistentPreRunE = nil

	// Map each command's cobra "SEE ALSO" filename (underscored path + .md) to
	// its final Starlight route, so cross-links land on the right nested page.
	links := map[string]string{}
	walk(root, func(c *cobra.Command) {
		links[linkKey(c)] = slugFor(c)
	})
	linkHandler := func(name string) string {
		if s, ok := links[name]; ok {
			return s
		}
		return "/reference/" + strings.ToLower(strings.TrimSuffix(name, ".md")) + "/"
	}

	var count int
	walk(root, func(c *cobra.Command) {
		c.DisableAutoGenTag = true
		if err := writePage(c, outDir, linkHandler); err != nil {
			log.Fatal(err)
		}
		count++
	})

	fmt.Printf("CLI reference generated in %s (%d pages)\n", outDir, count)
}

// writePage renders one command page with Starlight frontmatter into the
// directory that mirrors its position in the command tree.
func writePage(c *cobra.Command, outDir string, linkHandler func(string) string) error {
	var body bytes.Buffer
	if err := doc.GenMarkdownCustom(c, &body, linkHandler); err != nil {
		return err
	}

	var page strings.Builder
	page.WriteString("---\n")
	fmt.Fprintf(&page, "title: %s\n", yamlQuote(c.CommandPath()))
	if c.Short != "" {
		fmt.Fprintf(&page, "description: %s\n", yamlQuote(c.Short))
	}
	page.WriteString("sidebar:\n")
	fmt.Fprintf(&page, "  label: %s\n", yamlQuote(sidebarLabel(c)))
	if isOverview(c) {
		// Float a group's own page to the top of its group.
		page.WriteString("  order: 0\n")
	}
	page.WriteString("---\n\n")
	page.WriteString(stripLeadingHeading(body.String()))

	file := filepath.Join(outDir, filepath.FromSlash(relPath(c)))
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}
	return os.WriteFile(file, []byte(page.String()), 0644)
}

// pathParts returns the command path below the root, e.g. ["domains","records"].
func pathParts(c *cobra.Command) []string {
	parts := strings.Fields(c.CommandPath())
	return parts[1:]
}

func isOverview(c *cobra.Command) bool {
	return c.HasParent() && len(visibleSubcmds(c)) > 0
}

// relPath is the page's location relative to the reference root.
//   - root:           aic.md
//   - group (billing): billing/billing.md      (folder = group, file = overview)
//   - leaf (add-card): billing/add-card.md
func relPath(c *cobra.Command) string {
	parts := pathParts(c)
	if len(parts) == 0 {
		return "aic.md"
	}
	last := parts[len(parts)-1]
	if len(visibleSubcmds(c)) > 0 {
		return filepath.ToSlash(filepath.Join(append(parts, last+".md")...))
	}
	return filepath.ToSlash(filepath.Join(parts...)) + ".md"
}

func slugFor(c *cobra.Command) string {
	return "/reference/" + strings.TrimSuffix(relPath(c), ".md") + "/"
}

func sidebarLabel(c *cobra.Command) string {
	if !c.HasParent() {
		return c.Name()
	}
	if isOverview(c) {
		return "Overview"
	}
	return c.Name()
}

func linkKey(c *cobra.Command) string {
	return strings.ReplaceAll(c.CommandPath(), " ", "_") + ".md"
}

// walk visits a command and all visible descendants.
func walk(c *cobra.Command, fn func(*cobra.Command)) {
	fn(c)
	for _, sub := range visibleSubcmds(c) {
		walk(sub, fn)
	}
}

func visibleSubcmds(c *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, sub := range c.Commands() {
		if !sub.Hidden && sub.Name() != "help" && sub.Name() != "completion" {
			out = append(out, sub)
		}
	}
	return out
}

// stripLeadingHeading drops the leading "## ..." line (and the blank line after
// it) that cobra/doc emits; Starlight renders the frontmatter title as the H1.
func stripLeadingHeading(content string) string {
	line, rest, found := strings.Cut(content, "\n")
	if found && strings.HasPrefix(line, "## ") {
		return strings.TrimLeft(rest, "\n")
	}
	return content
}

func yamlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
