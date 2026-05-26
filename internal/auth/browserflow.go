// Package auth implements the browser-delegated start/open/poll flow shared by
// `aic login` and `aic billing add-card`.
package auth

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// ErrTimeout is returned when a flow does not complete before its deadline.
var ErrTimeout = errors.New("timed out waiting for browser flow to complete")

// Flow describes one browser-delegated flow. The function fields are injected
// so the flow is testable without a real browser or backend.
type Flow struct {
	Start       func(ctx context.Context) (sessionID, browserURL string, err error)
	OpenBrowser func(url string) error
	Poll        func(ctx context.Context, sessionID string) (status string, err error)
	Interval    time.Duration
	Timeout     time.Duration
}

// RunFlow starts the session, opens the browser, and polls until the status is
// terminal. Returns the session id on success.
func RunFlow(ctx context.Context, f Flow) (string, error) {
	id, browserURL, err := f.Start(ctx)
	if err != nil {
		return "", err
	}

	fmt.Println("Opening your browser to continue. If it does not open, visit:")
	fmt.Println("  " + browserURL)
	if err := f.OpenBrowser(browserURL); err != nil {
		// Non-fatal: the URL was printed above.
		fmt.Println("Could not open a browser automatically.")
	}

	deadline := time.Now().Add(f.Timeout)
	for {
		status, err := f.Poll(ctx, id)
		if err != nil {
			return "", err
		}
		switch status {
		case "completed":
			return id, nil
		case "expired":
			return "", errors.New("the session expired before completing")
		case "denied":
			return "", errors.New("the request was denied")
		}
		if time.Now().After(deadline) {
			return "", ErrTimeout
		}
		time.Sleep(f.Interval)
	}
}

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
