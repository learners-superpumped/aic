package cmd

import "testing"

func TestDomainsRecordsSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newDomainsRecordsCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"list", "add", "set", "delete", "import"} {
		if !names[want] {
			t.Errorf("missing records subcommand %q", want)
		}
	}
}

func TestRecordName(t *testing.T) {
	cases := []struct{ in, domain, want string }{
		{"@", "x.com", "x.com"},
		{"www", "x.com", "www.x.com"},
		{"www.x.com", "x.com", "www.x.com"},
	}
	for _, c := range cases {
		if got := recordName(c.in, c.domain); got != c.want {
			t.Errorf("recordName(%q,%q)=%q want %q", c.in, c.domain, got, c.want)
		}
	}
}
