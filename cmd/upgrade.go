package cmd

import "strings"

type upgradeAction int

const (
	actionUpgrade upgradeAction = iota // current is behind latest
	actionCurrent                      // already on latest
	actionDev                          // local/dev build, not upgradable
)

// decideUpgrade picks the action given the current and latest version strings.
// Comparison normalizes the leading "v"; equality means already-latest. Since
// releases/latest is by definition newest, any difference means upgrade.
func decideUpgrade(current, latest string) upgradeAction {
	if current == "dev" || current == "" {
		return actionDev
	}
	if strings.TrimPrefix(current, "v") == strings.TrimPrefix(latest, "v") {
		return actionCurrent
	}
	return actionUpgrade
}
