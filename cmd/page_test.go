package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/output"
	"github.com/spf13/cobra"
)

func TestPrintPageTableFooter(t *testing.T) {
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	cmd := &cobra.Command{Use: "list"}
	root := &cobra.Command{Use: "aic"}
	mid := &cobra.Command{Use: "ads"}
	root.AddCommand(mid)
	mid.AddCommand(cmd)

	page := api.Page[string]{Data: []string{"a", "b"}, HasMore: true, NextCursor: "CUR"}
	err := printPage(cmd, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} })
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("rows missing: %q", out)
	}
	if !strings.Contains(out, "aic ads list --cursor CUR") {
		t.Fatalf("footer missing: %q", out)
	}
}

func TestPrintPageTableFooterPreservesFilterFlags(t *testing.T) {
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	cmd := &cobra.Command{Use: "list"}
	root := &cobra.Command{Use: "aic"}
	mail := &cobra.Command{Use: "mail"}
	msgs := &cobra.Command{Use: "messages"}
	root.AddCommand(mail)
	mail.AddCommand(msgs)
	msgs.AddCommand(cmd)
	// Register the same local flags the real list command has, and mark the
	// filter ones as set (Changed), as cobra would after parsing.
	var direction, inbox, from, cursor string
	var limit int
	cmd.Flags().StringVar(&direction, "direction", "", "")
	cmd.Flags().StringVar(&inbox, "inbox", "", "")
	cmd.Flags().StringVar(&from, "from", "", "")
	cmd.Flags().IntVar(&limit, "limit", 0, "")
	cmd.Flags().StringVar(&cursor, "cursor", "", "")
	if err := cmd.Flags().Set("from", "alice@foo.com"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("limit", "2"); err != nil {
		t.Fatal(err)
	}

	page := api.Page[string]{Data: []string{"a", "b"}, HasMore: true, NextCursor: "CUR"}
	if err := printPage(cmd, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} }); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// The next-page command must carry the filter forward, else page 2 queries a
	// different result set than page 1.
	if !strings.Contains(out, "--from=alice@foo.com") {
		t.Fatalf("footer dropped --from: %q", out)
	}
	if !strings.Contains(out, "--limit=2") {
		t.Fatalf("footer dropped --limit: %q", out)
	}
	if !strings.Contains(out, "--cursor CUR") {
		t.Fatalf("footer missing fresh cursor: %q", out)
	}
	// An unset filter (--direction) must NOT leak into the command.
	if strings.Contains(out, "--direction") {
		t.Fatalf("footer leaked unset --direction: %q", out)
	}
}

func TestPrintPageTableFooterShellQuotesUnsafeValues(t *testing.T) {
	cases := []struct {
		val  string
		want string // expected substring in footer
	}{
		{"Acme Corp", `--name='Acme Corp'`}, // whitespace → single-quoted
		{"a$b@x.com", `--name='a$b@x.com'`}, // $ must be single-quoted so it can't expand
		{"a@x.com", `--name=a@x.com`},       // bare-safe → unquoted
		{"O'Brien", `--name='O'\''Brien'`},  // embedded single quote escaped
	}
	for _, c := range cases {
		var buf bytes.Buffer
		r, _ := output.New("table", &buf)
		cmd := &cobra.Command{Use: "list"}
		root := &cobra.Command{Use: "aic"}
		root.AddCommand(cmd)
		var name string
		cmd.Flags().StringVar(&name, "name", "", "")
		if err := cmd.Flags().Set("name", c.val); err != nil {
			t.Fatal(err)
		}
		page := api.Page[string]{Data: []string{"a"}, HasMore: true, NextCursor: "CUR"}
		if err := printPage(cmd, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} }); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), c.want) {
			t.Fatalf("value %q: want footer to contain %q, got %q", c.val, c.want, buf.String())
		}
	}
}

func TestPrintPageTableFooterScopeAndOutputGlobals(t *testing.T) {
	// Drive through cobra Execute so flag parsing + persistent-flag merge happen
	// exactly as in real CLI use — that's the only state where cmd.Flags().Visit
	// sees inherited globals as Changed. Scope overrides (--project/--team/
	// --profile) MUST carry forward — dropping them makes page 2 query a
	// different team/project than page 1. Presentation-only --output must not.
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	root := &cobra.Command{Use: "aic"}
	root.PersistentFlags().String("project", "", "")
	root.PersistentFlags().String("output", "", "")
	var from, cursor string
	var limit int
	list := &cobra.Command{Use: "list", RunE: func(c *cobra.Command, _ []string) error {
		page := api.Page[string]{Data: []string{"a"}, HasMore: true, NextCursor: "CUR"}
		return printPage(c, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} })
	}}
	list.Flags().StringVar(&from, "from", "", "")
	addPaginationFlags(list, &limit, &cursor)
	root.AddCommand(list)
	root.SetArgs([]string{"list", "--project", "proj_x", "--output", "table", "--from", "a@b.com", "--limit", "2"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "--from=a@b.com") || !strings.Contains(out, "--limit=2") {
		t.Fatalf("footer lost local filters: %q", out)
	}
	if !strings.Contains(out, "--project=proj_x") {
		t.Fatalf("footer dropped explicit scope override --project: %q", out)
	}
	if strings.Contains(out, "--output") {
		t.Fatalf("footer leaked presentation flag --output: %q", out)
	}
}

func TestPrintPageTableFooterPreservesPositionalArgs(t *testing.T) {
	// Commands like `storage ls <bucket>` pass their positional args through;
	// without them the suggested command fails cobra's Args validation.
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	root := &cobra.Command{Use: "aic"}
	ls := &cobra.Command{Use: "ls"}
	root.AddCommand(ls)
	page := api.Page[string]{Data: []string{"a"}, HasMore: true, NextCursor: "CUR"}
	if err := printPage(ls, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} }, "my bucket/prefix"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "aic ls 'my bucket/prefix' --cursor CUR") {
		t.Fatalf("footer dropped or misquoted positional arg: %q", buf.String())
	}
}

func TestPrintPageTableFooterBoolAndSliceFlags(t *testing.T) {
	// Bool flags must use --name=value (space-separated 'true' becomes a stray
	// positional on paste); slice flags must repeat --name=v per element
	// (Value.String()'s "[a,b]" doesn't re-parse).
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	cmd := &cobra.Command{Use: "list"}
	root := &cobra.Command{Use: "aic"}
	root.AddCommand(cmd)
	var archived bool
	var tags []string
	cmd.Flags().BoolVar(&archived, "archived", false, "")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "")
	if err := cmd.Flags().Set("archived", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("tag", "a"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("tag", "b c"); err != nil {
		t.Fatal(err)
	}
	page := api.Page[string]{Data: []string{"a"}, HasMore: true, NextCursor: "CUR"}
	if err := printPage(cmd, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} }); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "--archived=true") {
		t.Fatalf("bool flag not emitted as --name=value: %q", out)
	}
	if !strings.Contains(out, "--tag=a --tag='b c'") {
		t.Fatalf("slice flag not emitted per element: %q", out)
	}
}

func TestPrintPageTableNoFooter(t *testing.T) {
	var buf bytes.Buffer
	r, _ := output.New("table", &buf)
	cmd := &cobra.Command{Use: "list"}
	page := api.Page[string]{Data: []string{"a"}, HasMore: false}
	if err := printPage(cmd, r, page, []string{"V"}, func(v any) []string { return []string{v.(string)} }); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "Next page") {
		t.Fatalf("unexpected footer: %q", buf.String())
	}
}

func TestPrintPageJSONEnvelope(t *testing.T) {
	var buf bytes.Buffer
	r, _ := output.New("json", &buf)
	cmd := &cobra.Command{Use: "list"}
	page := api.Page[string]{Data: []string{"a"}, HasMore: true, NextCursor: "CUR"}
	if err := printPage(cmd, r, page, nil, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"has_more"`) || !strings.Contains(out, `"next_cursor"`) || !strings.Contains(out, "CUR") {
		t.Fatalf("envelope fields missing: %q", out)
	}
}
