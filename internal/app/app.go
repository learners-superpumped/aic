// Package app bundles per-invocation state (client, project, renderer) and
// carries it on a context.Context for commands to consume.
package app

import (
	"context"
	"errors"
	"io"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/output"
)

// Renderer is re-exported so commands depend only on app.
type Renderer = output.Renderer

// NewRenderer constructs a renderer (re-exported from output).
func NewRenderer(format string, w io.Writer) (*Renderer, error) {
	return output.New(format, w)
}

// App is everything a command needs at runtime.
type App struct {
	Client  *api.Client
	Project string
	Team    string
	Out     *Renderer
}

// RequireProject returns an error if no project is in scope.
func (a *App) RequireProject() error {
	if a.Project == "" {
		return errors.New("no project selected: pass --project or run `aic projects use <id>`")
	}
	return nil
}

// RequireTeam returns an error if no team is in scope.
func (a *App) RequireTeam() error {
	if a.Team == "" {
		return errors.New("no team selected: pass --team or run `aic teams switch <id>` or `aic login`")
	}
	return nil
}

type ctxKey struct{}

// NewContext stores a on ctx.
func NewContext(ctx context.Context, a *App) context.Context {
	return context.WithValue(ctx, ctxKey{}, a)
}

// FromContext retrieves the App, erroring if absent.
func FromContext(ctx context.Context) (*App, error) {
	a, ok := ctx.Value(ctxKey{}).(*App)
	if !ok || a == nil {
		return nil, errors.New("internal error: app not initialized")
	}
	return a, nil
}
