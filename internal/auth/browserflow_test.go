package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPollUntilCompleted(t *testing.T) {
	calls := 0
	opened := ""
	res, err := RunFlow(context.Background(), Flow{
		Start: func(ctx context.Context) (sessionID, browserURL string, err error) {
			return "s1", "https://x/login", nil
		},
		OpenBrowser: func(url string) error { opened = url; return nil },
		Poll: func(ctx context.Context, id string) (status string, err error) {
			calls++
			if calls < 3 {
				return "pending", nil
			}
			return "completed", nil
		},
		Interval: time.Millisecond,
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("RunFlow: %v", err)
	}
	if res != "s1" {
		t.Fatalf("want session id s1, got %q", res)
	}
	if opened != "https://x/login" {
		t.Fatalf("browser not opened to url, got %q", opened)
	}
	if calls < 3 {
		t.Fatalf("expected polling, calls=%d", calls)
	}
}

func TestPollDenied(t *testing.T) {
	_, err := RunFlow(context.Background(), Flow{
		Start:       func(ctx context.Context) (string, string, error) { return "s1", "u", nil },
		OpenBrowser: func(string) error { return nil },
		Poll:        func(ctx context.Context, id string) (string, error) { return "denied", nil },
		Interval:    time.Millisecond,
		Timeout:     time.Second,
	})
	if err == nil {
		t.Fatal("expected error on denied status")
	}
}

func TestPollTimeout(t *testing.T) {
	_, err := RunFlow(context.Background(), Flow{
		Start:       func(ctx context.Context) (string, string, error) { return "s1", "u", nil },
		OpenBrowser: func(string) error { return nil },
		Poll:        func(ctx context.Context, id string) (string, error) { return "pending", nil },
		Interval:    5 * time.Millisecond,
		Timeout:     20 * time.Millisecond,
	})
	if err == nil || !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}
