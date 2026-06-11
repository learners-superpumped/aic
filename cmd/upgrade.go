package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/learners-superpumped/aic/internal/selfupdate"
	"github.com/spf13/cobra"
)

type upgradeAction int

const (
	actionUpgrade upgradeAction = iota // current is behind latest
	actionCurrent                      // already on latest
	actionDev                          // local/dev build, not upgradable
)

var resolveExecPath = func() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved, nil
	}
	return p, nil
}

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Update aic to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if Version == "dev" || Version == "" {
				fmt.Fprintln(out, "This is a local/dev build of aic and can't be upgraded. Install a release first.")
				return nil
			}

			execPath, err := resolveExecPath()
			if err != nil {
				return err
			}
			if selfupdate.Method(execPath) == selfupdate.Managed {
				fmt.Fprintln(out, "aic was installed with npm. Update it with:")
				fmt.Fprintln(out, "  npm update -g @runaic/aic")
				return nil
			}

			latest, err := selfupdate.LatestVersion(cmd.Context())
			if err != nil {
				return err
			}

			switch decideUpgrade(Version, latest) {
			case actionCurrent:
				fmt.Fprintf(out, "aic is already on the latest version (%s).\n", latest)
				return nil
			case actionUpgrade:
				fmt.Fprintf(out, "Upgrading aic %s → %s…\n", Version, latest)
				installDir := filepath.Dir(execPath)
				return selfupdate.RunInstaller(cmd.Context(), latest, installDir, out)
			default:
				return nil
			}
		},
	}
}

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
