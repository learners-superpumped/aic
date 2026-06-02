// Command docs generates the aic CLI command reference as Starlight-flavored
// Markdown, one page per command, with YAML frontmatter the docs site consumes.
//
//	go run ./tools/docs [outDir]   # default: ./reference-out
package main

import (
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
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := cleanMarkdown(outDir); err != nil {
		log.Fatal(err)
	}

	root := cmd.NewRootCmd()
	root.DisableAutoGenTag = true
	root.PersistentPreRunE = nil

	// Cobra "SEE ALSO" links reference sibling pages by filename
	// (e.g. aic_teams.md). Map them to Starlight reference routes.
	linkHandler := func(name string) string {
		return "/reference/" + strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name))) + "/"
	}
	filePrepender := func(string) string { return "" }

	if err := doc.GenMarkdownTreeCustom(root, outDir, filePrepender, linkHandler); err != nil {
		log.Fatal(err)
	}

	if err := addFrontmatter(root, outDir); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CLI reference generated in %s\n", outDir)
}

// cleanMarkdown removes previously generated pages so renamed/removed commands
// don't leave stale files behind, while preserving .gitkeep.
func cleanMarkdown(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			return err
		}
	}
	return nil
}

// addFrontmatter rewrites each generated page: it injects a Starlight YAML
// frontmatter block (title = command path, description = Short) and strips the
// leading "## <command path>" heading that Starlight already renders as the H1.
func addFrontmatter(root *cobra.Command, dir string) error {
	shorts := map[string]string{}
	collectShorts(root, shorts)

	matches, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		title := strings.ReplaceAll(strings.TrimSuffix(filepath.Base(path), ".md"), "_", " ")
		body := stripLeadingHeading(string(raw))

		var fm strings.Builder
		fm.WriteString("---\n")
		fmt.Fprintf(&fm, "title: %s\n", yamlQuote(title))
		if s := shorts[title]; s != "" {
			fmt.Fprintf(&fm, "description: %s\n", yamlQuote(s))
		}
		fm.WriteString("---\n\n")
		fm.WriteString(body)

		if err := os.WriteFile(path, []byte(fm.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

// collectShorts maps every command's full path ("aic teams list") to its Short.
func collectShorts(c *cobra.Command, out map[string]string) {
	out[c.CommandPath()] = c.Short
	for _, sub := range c.Commands() {
		collectShorts(sub, out)
	}
}

// stripLeadingHeading drops the leading "## ..." line (and the blank line after
// it) that cobra/doc emits as the page heading.
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
