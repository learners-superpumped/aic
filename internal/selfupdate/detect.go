package selfupdate

import (
	"path/filepath"
	"strings"
)

// Install describes how the running binary was installed.
type Install int

const (
	Direct  Install = iota // installed via install.sh; self-update allowed
	Managed                // installed under a package manager (npm); defer to it
)

// Method classifies an install from the resolved executable path. A
// "node_modules" path segment means an npm-managed install.
func Method(execPath string) Install {
	for _, seg := range strings.Split(filepath.ToSlash(execPath), "/") {
		if seg == "node_modules" {
			return Managed
		}
	}
	return Direct
}
