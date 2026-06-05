package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/learners-superpumped/aic/internal/output"
)

func TestSEOSite_RendersAllFormats(t *testing.T) {
	site := api.SEOSiteDTO{Domain: "example.com", DNSManaged: true, Status: "active"}
	headers := []string{"DOMAIN", "DNS", "STATUS"}
	rowFn := func(v any) []string {
		s := v.(api.SEOSiteDTO)
		return []string{s.Domain, boolYN(s.DNSManaged), s.Status}
	}

	for _, format := range []string{"table", "json", "yaml"} {
		var buf bytes.Buffer
		r, err := output.New(format, &buf)
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		if err := r.Print(site, headers, rowFn); err != nil {
			t.Fatalf("%s print: %v", format, err)
		}
		if buf.Len() == 0 {
			t.Fatalf("%s produced no output", format)
		}
	}

	// json round-trips back to the DTO.
	var buf bytes.Buffer
	r, _ := output.New("json", &buf)
	_ = r.Print(site, headers, rowFn)
	var back api.SEOSiteDTO
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil || back.Domain != "example.com" || !back.DNSManaged {
		t.Fatalf("json round-trip: %+v err=%v", back, err)
	}
}
