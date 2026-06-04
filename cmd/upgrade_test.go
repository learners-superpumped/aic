package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpgradeCmd_DevBuildRefuses(t *testing.T) {
	var buf bytes.Buffer
	cmd := newUpgradeCmd()
	cmd.SetOut(&buf)
	cmd.SetContext(t.Context())

	prev := Version
	Version = "dev"
	defer func() { Version = prev }()

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("dev build should not error: %v", err)
	}
	if !strings.Contains(buf.String(), "dev") && !strings.Contains(buf.String(), "local") {
		t.Errorf("expected dev-build guidance, got: %s", buf.String())
	}
}

func TestUpgradeCmd_NpmManagedGuides(t *testing.T) {
	var buf bytes.Buffer
	cmd := newUpgradeCmd()
	cmd.SetOut(&buf)
	cmd.SetContext(t.Context())

	prevV := Version
	Version = "v1.4.0"
	defer func() { Version = prevV }()

	prevR := resolveExecPath
	resolveExecPath = func() (string, error) {
		return "/usr/lib/node_modules/@runaic/aic/bin/aic", nil
	}
	defer func() { resolveExecPath = prevR }()

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("npm-managed should not error: %v", err)
	}
	if !strings.Contains(buf.String(), "npm update -g @runaic/aic") {
		t.Errorf("expected npm guidance, got: %s", buf.String())
	}
}

func TestUpgradeDecision(t *testing.T) {
	cases := []struct {
		name, current, latest string
		want                  upgradeAction
	}{
		{"dev build", "dev", "v1.5.0", actionDev},
		{"already latest", "v1.5.0", "v1.5.0", actionCurrent},
		{"already latest no-v", "1.5.0", "v1.5.0", actionCurrent},
		{"older", "v1.4.0", "v1.5.0", actionUpgrade},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := decideUpgrade(c.current, c.latest); got != c.want {
				t.Errorf("decideUpgrade(%q,%q) = %v, want %v", c.current, c.latest, got, c.want)
			}
		})
	}
}
