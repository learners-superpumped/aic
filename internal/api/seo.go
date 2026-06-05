package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type SEOSiteDTO struct {
	Domain            string `json:"domain"`
	DNSManaged        bool   `json:"dns_managed"`
	Status            string `json:"status"`
	VerifyRecordName  string `json:"verify_record_name,omitempty"`
	VerifyRecordValue string `json:"verify_record_value,omitempty"`
}

type SCMetricDTO struct {
	Date        string            `json:"date"`
	Dimensions  map[string]string `json:"dimensions,omitempty"`
	Clicks      int64             `json:"clicks"`
	Impressions int64             `json:"impressions"`
	CTR         float64           `json:"ctr"`
	Position    float64           `json:"position"`
}

type SCSitemapDTO struct {
	Path          string `json:"path"`
	LastSubmitted string `json:"last_submitted"`
	IsPending     bool   `json:"is_pending"`
	Warnings      int64  `json:"warnings"`
	Errors        int64  `json:"errors"`
}

type SCInspectionDTO struct {
	URL           string `json:"url"`
	CoverageState string `json:"coverage_state"`
	IndexingState string `json:"indexing_state"`
	LastCrawlTime string `json:"last_crawl_time"`
	Verdict       string `json:"verdict"`
}

func seoBase(team, project string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/seo", url.PathEscape(team), url.PathEscape(project))
}

func (c *Client) AddSEOSite(ctx context.Context, team, project, domain string) (SEOSiteDTO, error) {
	var out SEOSiteDTO
	return out, c.do(ctx, "POST", seoBase(team, project)+"/sites", map[string]string{"domain": domain}, &out)
}

func (c *Client) ListSEOSites(ctx context.Context, team, project string, limit int, cursor string) (Page[SEOSiteDTO], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := seoBase(team, project) + "/sites"
	if e := q.Encode(); e != "" {
		path += "?" + e
	}
	var out Page[SEOSiteDTO]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) ShowSEOSite(ctx context.Context, team, project, domain string) (SEOSiteDTO, error) {
	var out SEOSiteDTO
	return out, c.do(ctx, "GET", seoBase(team, project)+"/sites/"+url.PathEscape(domain), nil, &out)
}

func (c *Client) VerifySEOSite(ctx context.Context, team, project, domain string) (string, error) {
	var out struct {
		Status string `json:"status"`
	}
	err := c.do(ctx, "POST", seoBase(team, project)+"/sites/"+url.PathEscape(domain)+"/verify", nil, &out)
	return out.Status, err
}

func (c *Client) DeleteSEOSite(ctx context.Context, team, project, domain string) error {
	return c.do(ctx, "DELETE", seoBase(team, project)+"/sites/"+url.PathEscape(domain), nil, nil)
}

type SCFilter struct {
	Dimension  string `json:"dimension"`
	Operator   string `json:"operator"`
	Expression string `json:"expression"`
}

type SCQuery struct {
	StartDate  string     `json:"start_date,omitempty"`
	EndDate    string     `json:"end_date,omitempty"`
	Dimensions []string   `json:"dimensions,omitempty"`
	Filters    []SCFilter `json:"filters,omitempty"`
	Type       string     `json:"type,omitempty"`
	StartRow   int        `json:"start_row,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

func (c *Client) SCQuery(ctx context.Context, team, project, domain string, q SCQuery) ([]SCMetricDTO, error) {
	var out []SCMetricDTO
	return out, c.do(ctx, "POST", seoBase(team, project)+"/sites/"+url.PathEscape(domain)+"/search-console/query", q, &out)
}

func (c *Client) SCSitemaps(ctx context.Context, team, project, domain string) ([]SCSitemapDTO, error) {
	var out []SCSitemapDTO
	return out, c.do(ctx, "GET", seoBase(team, project)+"/sites/"+url.PathEscape(domain)+"/search-console/sitemaps", nil, &out)
}

func (c *Client) SCSubmitSitemap(ctx context.Context, team, project, domain, sitemapURL string) error {
	return c.do(ctx, "POST", seoBase(team, project)+"/sites/"+url.PathEscape(domain)+"/search-console/sitemaps",
		map[string]string{"url": sitemapURL}, nil)
}

func (c *Client) SCDeleteSitemap(ctx context.Context, team, project, domain, sitemapURL string) error {
	path := seoBase(team, project) + "/sites/" + url.PathEscape(domain) + "/search-console/sitemaps?url=" + url.QueryEscape(sitemapURL)
	return c.do(ctx, "DELETE", path, nil, nil)
}

func (c *Client) SCInspect(ctx context.Context, team, project, domain, target string) (SCInspectionDTO, error) {
	var out SCInspectionDTO
	return out, c.do(ctx, "POST", seoBase(team, project)+"/sites/"+url.PathEscape(domain)+"/search-console/inspect", map[string]string{"url": target}, &out)
}
