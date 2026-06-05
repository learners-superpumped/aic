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
