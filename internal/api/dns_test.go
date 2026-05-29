package api

import (
	"encoding/json"
	"strings"
	"testing"
)

// The server rejects unknown fields on add/set bodies. The DNSRecord type is
// shared with responses (which carry source), so a request must omit source.
func TestDNSRecord_RequestOmitsSource(t *testing.T) {
	b, err := json.Marshal(DNSRecord{Type: "A", Name: "x.example.com", Values: []string{"1.2.3.4"}, TTL: 300})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "source") {
		t.Fatalf("add/set body must not contain source, got %s", b)
	}
}
