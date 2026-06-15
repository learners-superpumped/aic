package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/app"
)

func TestAdsCmdHasSubcommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newAdsCmd().Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"launch", "list", "status", "pause", "resume", "delete"} {
		if !names[want] {
			t.Errorf("missing ads subcommand %q", want)
		}
	}
}

func TestAdsLaunchRequiresProject(t *testing.T) {
	a := newAppNoProject(t)
	launch := findSub(newAdsCmd(), "launch")
	launch.SetContext(ctxWithApp(t, a))
	if err := launch.RunE(launch, nil); err == nil {
		t.Fatal("expected error when no project selected")
	}
}

func TestAdsLaunchCallsAPIAndOutputsCampaign(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/t1/projects/p1/ads/campaigns" {
			t.Errorf("want POST .../ads/campaigns, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"cmp_1","team_id":"t1","project_id":"p1","provider":"meta","objective":"traffic","budget_type":"daily","budget_nano":10000000000,"status":"submitted","launch_token":"tok","created_at":"2026-06-04T00:00:00Z","updated_at":"2026-06-04T00:00:00Z"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	launch := findSub(newAdsCmd(), "launch")
	launch.SetContext(ctxWithApp(t, a))
	launch.Flags().Set("objective", "traffic")
	launch.Flags().Set("budget-type", "daily")
	launch.Flags().Set("budget", "10000000000")
	launch.Flags().Set("headline", "Try AIC")
	launch.Flags().Set("body", "The platform.")
	launch.Flags().Set("cta", "Learn More")
	launch.Flags().Set("url", "https://runaic.com")

	if err := launch.RunE(launch, nil); err != nil {
		t.Fatalf("launch: %v", err)
	}
	if !strings.Contains(buf.String(), "cmp_1") {
		t.Fatalf("expected campaign id in output: %s", buf.String())
	}
}

func TestAdsLaunchAgeRangeParsed(t *testing.T) {
	ageMin, ageMax, err := parseAgeRange("25-44")
	if err != nil {
		t.Fatalf("parseAgeRange: %v", err)
	}
	if ageMin != 25 || ageMax != 44 {
		t.Errorf("want 25/44, got %d/%d", ageMin, ageMax)
	}
}

func TestAdsLaunchAgeRangeInvalidRejected(t *testing.T) {
	for _, bad := range []string{"", "25", "-44", "abc-def", "25-abc", "-"} {
		if _, _, err := parseAgeRange(bad); err == nil {
			t.Errorf("parseAgeRange(%q) expected error, got nil", bad)
		}
	}
}

// An open-ended upper bound ("24-" / "24+") yields max 0 so the targeting omits
// age_max — required for Advantage+ audience, which rejects a hard maximum age.
func TestAdsLaunchAgeRangeOpenUpperBound(t *testing.T) {
	for _, in := range []string{"24-", "24+", " 24- "} {
		mn, mx, err := parseAgeRange(in)
		if err != nil {
			t.Fatalf("parseAgeRange(%q): %v", in, err)
		}
		if mn != 24 || mx != 0 {
			t.Errorf("parseAgeRange(%q) = %d/%d, want 24/0", in, mn, mx)
		}
	}
}

func TestAdsListHitsProjectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"data":[{"id":"cmp_1","status":"active"}],"has_more":false}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	list := findSub(newAdsCmd(), "list")
	list.SetContext(ctxWithApp(t, a))
	if err := list.RunE(list, nil); err != nil {
		t.Fatalf("list: %v", err)
	}
	if gotPath != "/v1/teams/t1/projects/p1/ads/campaigns" {
		t.Errorf("path mismatch: %s", gotPath)
	}
	if !strings.Contains(buf.String(), "cmp_1") {
		t.Fatalf("expected cmp_1 in output: %s", buf.String())
	}
}

func TestAdsPauseHitsCorrectPath(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	pause := findSub(newAdsCmd(), "pause")
	pause.SetContext(ctxWithApp(t, a))
	if err := pause.RunE(pause, []string{"cmp_1"}); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/teams/t1/projects/p1/ads/campaigns/cmp_1/pause" {
		t.Errorf("want POST .../pause, got %s %s", gotMethod, gotPath)
	}
}

func TestAdsPixelCreateOutputsSnippet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/teams/t1/projects/p1/ads/pixel" {
			t.Errorf("want POST .../ads/pixel, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"project_id":"p1","pixel_id":"px_abc123","status":"active","created_at":"2026-06-11T00:00:00Z"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	pixelCmd := findSub(newAdsCmd(), "pixel")
	create := findSub(pixelCmd, "create")
	create.SetContext(ctxWithApp(t, a))
	create.SetOut(&buf)
	if err := create.RunE(create, nil); err != nil {
		t.Fatalf("pixel create: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "px_abc123") {
		t.Errorf("expected pixel id in output: %s", out)
	}
	if !strings.Contains(out, "fbq('init'") {
		t.Errorf("expected fbq('init' snippet in output: %s", out)
	}
}

func TestAdsPixelCreateJSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"project_id":"p1","pixel_id":"px_json1","status":"active","created_at":"2026-06-11T00:00:00Z"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	pixelCmd := findSub(newAdsCmd(), "pixel")
	create := findSub(pixelCmd, "create")
	create.SetContext(ctxWithApp(t, a))
	if err := create.RunE(create, nil); err != nil {
		t.Fatalf("pixel create json: %v", err)
	}
	if !strings.Contains(buf.String(), "px_json1") {
		t.Fatalf("expected pixel id in json output: %s", buf.String())
	}
}

func TestAdsPixelStatusReceivingEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/teams/t1/projects/p1/ads/pixel" {
			t.Errorf("want GET .../ads/pixel, got %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"project_id":"p1","pixel_id":"px_abc123","status":"active","created_at":"2026-06-11T00:00:00Z","stats":{"has_recent_events":true,"last_fired_at":"2026-06-11T01:00:00Z"}}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("table", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	pixelCmd := findSub(newAdsCmd(), "pixel")
	status := findSub(pixelCmd, "status")
	status.SetContext(ctxWithApp(t, a))
	status.SetOut(&buf)
	if err := status.RunE(status, nil); err != nil {
		t.Fatalf("pixel status: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "receiving events") {
		t.Errorf("expected 'receiving events' in output: %s", out)
	}
}

func TestAdsPixelListHitsTeamPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"data":[{"pixel_id":"px_1","project_id":"p1","status":"active","created_at":"2026-06-11T00:00:00Z"}],"has_more":false}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	pixelCmd := findSub(newAdsCmd(), "pixel")
	list := findSub(pixelCmd, "list")
	list.SetContext(ctxWithApp(t, a))
	if err := list.RunE(list, nil); err != nil {
		t.Fatalf("pixel list: %v", err)
	}
	if gotPath != "/v1/teams/t1/ads/pixels" {
		t.Errorf("want /v1/teams/t1/ads/pixels, got %s", gotPath)
	}
	if !strings.Contains(buf.String(), "px_1") {
		t.Fatalf("expected px_1 in output: %s", buf.String())
	}
}

func TestAdsLaunchConversionEvent(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"cmp_2","status":"submitted","objective":"conversions","budget_type":"daily","budget_nano":5000000000,"created_at":"2026-06-11T00:00:00Z","updated_at":"2026-06-11T00:00:00Z"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	r, _ := app.NewRenderer("json", &buf)
	a := &app.App{Client: api.New(srv.URL, "tok"), Team: "t1", Project: "p1", Out: r}

	launch := findSub(newAdsCmd(), "launch")
	launch.SetContext(ctxWithApp(t, a))
	launch.Flags().Set("objective", "conversions")
	launch.Flags().Set("budget-type", "daily")
	launch.Flags().Set("budget", "5000000000")
	launch.Flags().Set("conversion-event", "Purchase")

	if err := launch.RunE(launch, nil); err != nil {
		t.Fatalf("launch with conversion-event: %v", err)
	}
	if !strings.Contains(string(body), `"conversion_event":"Purchase"`) {
		t.Errorf("expected conversion_event in request body: %s", body)
	}
}
