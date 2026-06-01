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
	outDir := "./wiki-out"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}

	root := cmd.NewRootCmd()
	root.DisableAutoGenTag = true
	root.PersistentPreRunE = nil

	linkHandler := func(name string) string {
		return strings.TrimSuffix(name, filepath.Ext(name))
	}
	filePrepender := func(_ string) string { return "" }

	if err := doc.GenMarkdownTreeCustom(root, outDir, filePrepender, linkHandler); err != nil {
		log.Fatal(err)
	}

	if err := writeHome(outDir, root); err != nil {
		log.Fatal(err)
	}
	if err := writeSidebar(outDir, root); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("docs generated in %s\n", outDir)
}

func writeHome(outDir string, root *cobra.Command) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# aic CLI Reference\n\n%s\n\n## Commands\n\n", root.Short)

	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		page := strings.ReplaceAll(c.CommandPath(), " ", "_")
		fmt.Fprintf(&sb, "### [[%s|%s]]\n\n%s\n\n", page, c.Name(), c.Short)
		for _, sub := range visibleSubcmds(c) {
			subPage := strings.ReplaceAll(sub.CommandPath(), " ", "_")
			fmt.Fprintf(&sb, "- [[%s|%s %s]] — %s\n", subPage, c.Name(), sub.Name(), sub.Short)
		}
		if len(visibleSubcmds(c)) > 0 {
			sb.WriteByte('\n')
		}
	}
	return os.WriteFile(filepath.Join(outDir, "Home.md"), []byte(sb.String()), 0644)
}

func writeSidebar(outDir string, root *cobra.Command) error {
	var sb strings.Builder
	sb.WriteString("**[Home](Home)**\n\n")

	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		page := strings.ReplaceAll(c.CommandPath(), " ", "_")
		fmt.Fprintf(&sb, "**[[%s|%s]]**\n", page, c.Name())
		for _, sub := range visibleSubcmds(c) {
			subPage := strings.ReplaceAll(sub.CommandPath(), " ", "_")
			fmt.Fprintf(&sb, "- [[%s|%s]]\n", subPage, sub.Name())
			for _, leaf := range visibleSubcmds(sub) {
				leafPage := strings.ReplaceAll(leaf.CommandPath(), " ", "_")
				fmt.Fprintf(&sb, "  - [[%s|%s]]\n", leafPage, leaf.Name())
			}
		}
		sb.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(outDir, "_Sidebar.md"), []byte(sb.String()), 0644)
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
