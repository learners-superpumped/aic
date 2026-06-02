package output

import (
	"bytes"
	"strings"
	"testing"
)

type row struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	r, err := New("table", &buf)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Print([]row{{"1", "alpha"}, {"2", "beta"}}, []string{"ID", "Name"}, func(v any) []string {
		x := v.(row)
		return []string{x.ID, x.Name}
	}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("table missing content:\n%s", out)
	}
}

func TestTableSingleValue(t *testing.T) {
	var buf bytes.Buffer
	r, _ := New("table", &buf)
	if err := r.Print(row{"1", "alpha"}, []string{"ID", "Name"}, func(v any) []string {
		x := v.(row)
		return []string{x.ID, x.Name}
	}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "alpha") {
		t.Fatalf("single-value table missing content:\n%s", out)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	r, _ := New("json", &buf)
	if err := r.Print([]row{{"1", "alpha"}}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"name": "alpha"`) {
		t.Fatalf("json missing field:\n%s", buf.String())
	}
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	r, _ := New("yaml", &buf)
	if err := r.Print(row{"1", "alpha"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "name: alpha") {
		t.Fatalf("yaml missing field:\n%s", buf.String())
	}
}

// TestYAMLHonorsJSONTags locks in the DRY behavior: YAML output honors `json`
// tags and flattens embedded structs (encoded via JSON), so DTOs need no
// `yaml` tags and YAML matches JSON exactly.
func TestYAMLHonorsJSONTags(t *testing.T) {
	type inner struct {
		MailFrom string `json:"mail_from"`
	}
	type outer struct {
		inner
		ReplyTo string `json:"reply_to"`
	}
	var buf bytes.Buffer
	r, _ := New("yaml", &buf)
	if err := r.Print(outer{inner{"a@b.com"}, "c@d.com"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "mail_from: a@b.com") || !strings.Contains(out, "reply_to: c@d.com") {
		t.Fatalf("yaml did not honor json tags:\n%s", out)
	}
	if strings.Contains(out, "inner:") {
		t.Fatalf("embedded struct not flattened:\n%s", out)
	}
}

// TestTableCJKAlignment locks in display-width column alignment: a CJK cell is
// twice as wide on screen as text/tabwriter's rune count assumes, so the next
// column must still start at the same screen offset across rows.
func TestTableCJKAlignment(t *testing.T) {
	var buf bytes.Buffer
	r, _ := New("table", &buf)
	if err := r.Print(
		[]row{{"1", "답장 테스트"}, {"2", "ascii"}},
		[]string{"ID", "Name"},
		func(v any) []string { x := v.(row); return []string{x.ID, x.Name} },
	); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d:\n%s", len(lines), buf.String())
	}
	// In each row the second column ("Name") starts after the ID cell padded to
	// the ID column's display width. ID cells are all width 1 here, so the Name
	// column must begin at the same screen offset on every line.
	off := func(line string) int { return displayWidth(line[:strings.IndexAny(line, "답a")]) }
	if a, b := off(lines[1]), off(lines[2]); a != b {
		t.Fatalf("Name column misaligned: row1 starts at %d, row2 at %d:\n%s", a, b, buf.String())
	}
}

func TestDisplayWidth(t *testing.T) {
	if w := displayWidth("ab"); w != 2 {
		t.Fatalf("ascii width=%d, want 2", w)
	}
	if w := displayWidth("한글"); w != 4 {
		t.Fatalf("CJK width=%d, want 4", w)
	}
}

func TestInvalidFormat(t *testing.T) {
	if _, err := New("xml", &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for unknown format")
	}
}
