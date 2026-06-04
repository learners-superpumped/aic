package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// installScriptURL is the canonical install script; overridable in tests.
var installScriptURL = "https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh"

// RunInstaller downloads install.sh and pipes it to `sh`, instructing it to
// install the given version into installDir. install.sh owns checksum
// verification and sudo handling. Output is streamed to out.
func RunInstaller(ctx context.Context, version, installDir string, out io.Writer) error {
	script := fmt.Sprintf("curl -fsSL %s | sh", installScriptURL)
	cmd := exec.CommandContext(ctx, "sh", "-c", script)
	cmd.Env = append(os.Environ(),
		"AIC_VERSION="+version,
		"AIC_INSTALL_DIR="+installDir,
	)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}
	return nil
}
