package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aic/internal/app"
)

// actionResult is the structured form of a mutation that returns no object
// (delete, set-default, ...). In json/yaml it carries the affected resource's
// minimal identity plus a status verb; table mode prints a human line instead.
type actionResult struct {
	Name   string `json:"name" yaml:"name"`
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	Status string `json:"status" yaml:"status"`
}

// printAction renders a no-object mutation: the structured result for
// json/yaml, or tableLine verbatim for table mode.
func printAction(a *app.App, res actionResult, tableLine string) error {
	if a.Out.Format() != "table" {
		return a.Out.Print(res, nil, nil)
	}
	_, err := fmt.Fprintln(a.Out.Writer(), tableLine)
	return err
}
