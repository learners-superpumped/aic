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

func TestInvalidFormat(t *testing.T) {
	if _, err := New("xml", &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for unknown format")
	}
}
