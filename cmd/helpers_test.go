package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/learners-company/aic/internal/api"
	"github.com/learners-company/aic/internal/app"
	"github.com/spf13/cobra"
)

func newAppNoProject(t *testing.T) *app.App {
	t.Helper()
	r, _ := app.NewRenderer("json", &bytes.Buffer{})
	return &app.App{Client: api.New("http://unused", "tok"), Project: "", Out: r}
}

func ctxWithApp(t *testing.T, a *app.App) context.Context {
	t.Helper()
	return app.NewContext(context.Background(), a)
}

func findSub(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
