package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var installScriptURL = "https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh"

// RunInstaller pipes install.sh to sh to install version into installDir.
// install.sh owns checksum verification and sudo handling, so we reuse it
// rather than reimplement those here.
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
