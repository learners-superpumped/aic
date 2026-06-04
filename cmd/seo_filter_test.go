package cmd

import "testing"

func TestParseFilter(t *testing.T) {
	f, err := parseFilter("country equals usa")
	if err != nil || f.Dimension != "country" || f.Operator != "equals" || f.Expression != "usa" {
		t.Fatalf("simple: %+v err=%v", f, err)
	}
	// expression may contain spaces / slashes
	f, err = parseFilter("page contains /blog/post 1")
	if err != nil || f.Dimension != "page" || f.Operator != "contains" || f.Expression != "/blog/post 1" {
		t.Fatalf("spaced expr: %+v err=%v", f, err)
	}
	if _, err := parseFilter("bad"); err == nil {
		t.Fatal("expected error for <3 parts")
	}
}
