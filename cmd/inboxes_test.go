package cmd

import (
	"testing"
)

func TestInboxesCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newInboxesCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"create", "list", "delete", "show"} {
		if !names[want] {
			t.Errorf("missing inboxes subcommand %q", want)
		}
	}
}

func TestInboxesCreateRequiresProject(t *testing.T) {
	a := newAppNoProject(t)
	create := findSub(newInboxesCmd(), "create")
	create.SetContext(ctxWithApp(t, a))
	if err := create.RunE(create, []string{"a@x.com"}); err == nil {
		t.Fatal("expected error when no project selected")
	}
}
