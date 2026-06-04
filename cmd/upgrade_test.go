package cmd

import "testing"

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
