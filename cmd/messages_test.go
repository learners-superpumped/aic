package cmd

import (
	"testing"
)

func TestMessagesCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newMessagesCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"send", "list", "show"} {
		if !names[want] {
			t.Errorf("missing messages subcommand %q", want)
		}
	}
}

func TestMessagesSendRequiresProject(t *testing.T) {
	a := newAppNoProject(t)
	send := findSub(newMessagesCmd(), "send")
	send.Flags().Set("inbox", "a@x.com")
	send.Flags().Set("to", "b@y.com")
	send.SetContext(ctxWithApp(t, a))
	if err := send.RunE(send, nil); err == nil {
		t.Fatal("expected error when no project selected")
	}
}
