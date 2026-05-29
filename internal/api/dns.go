package api

import (
	"context"
	"net/http"
	"net/url"
)

type DNSRecord struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values"`
	TTL    int32    `json:"ttl"`
	Source string   `json:"source,omitempty"` // response-only; omitted from add/set bodies (server rejects unknown fields)
}

type dnsRecordsResponse struct {
	Records  []DNSRecord `json:"records"`
	Warnings []string    `json:"warnings"`
}

func dnsRecordsPath(teamID, projectID, domain string) string {
	return "/v1/teams/" + url.PathEscape(teamID) +
		"/projects/" + url.PathEscape(projectID) +
		"/domains/" + url.PathEscape(domain) + "/records"
}

func (c *Client) ListDNSRecords(ctx context.Context, teamID, projectID, domain string) ([]DNSRecord, error) {
	var out dnsRecordsResponse
	err := c.do(ctx, http.MethodGet, dnsRecordsPath(teamID, projectID, domain), nil, &out)
	return out.Records, err
}

func (c *Client) AddDNSRecord(ctx context.Context, teamID, projectID, domain string, r DNSRecord) (*DNSRecord, error) {
	var out DNSRecord
	return &out, c.do(ctx, http.MethodPost, dnsRecordsPath(teamID, projectID, domain), r, &out)
}

func (c *Client) SetDNSRecord(ctx context.Context, teamID, projectID, domain string, r DNSRecord) (*DNSRecord, error) {
	var out DNSRecord
	return &out, c.do(ctx, http.MethodPut, dnsRecordsPath(teamID, projectID, domain), r, &out)
}

func (c *Client) DeleteDNSRecord(ctx context.Context, teamID, projectID, domain, name, recordType string) error {
	p := dnsRecordsPath(teamID, projectID, domain) + "?name=" + url.QueryEscape(name) + "&type=" + url.QueryEscape(recordType)
	return c.do(ctx, http.MethodDelete, p, nil, nil)
}

func (c *Client) ImportDNSRecords(ctx context.Context, teamID, projectID, domain string) ([]DNSRecord, []string, error) {
	var out dnsRecordsResponse
	err := c.do(ctx, http.MethodPost, dnsRecordsPath(teamID, projectID, domain)+"/import", nil, &out)
	return out.Records, out.Warnings, err
}
