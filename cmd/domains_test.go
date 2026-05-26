package cmd

import (
	"testing"
)

func TestDomainsCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newDomainsCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"search", "buy", "list", "show"} {
		if !names[want] {
			t.Errorf("missing domains subcommand %q", want)
		}
	}
}

func TestDomainsBuyRequiresProject(t *testing.T) {
	a := newAppNoProject(t)
	buy := findSub(newDomainsCmd(), "buy")
	buy.SetContext(ctxWithApp(t, a))
	err := buy.RunE(buy, []string{"example.com"})
	if err == nil {
		t.Fatal("expected error when no project selected")
	}
}
