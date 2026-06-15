package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func adsBasePath(teamID, projectID string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/ads",
		url.PathEscape(teamID), url.PathEscape(projectID))
}

func (c *Client) LaunchAd(ctx context.Context, teamID, projectID string, req AdLaunchRequest) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/campaigns", req, &out)
}

func (c *Client) ListAds(ctx context.Context, teamID, projectID string, statuses []string, limit int, cursor string) (Page[AdCampaign], error) {
	var extra url.Values
	if len(statuses) > 0 {
		extra = url.Values{"status": {strings.Join(statuses, ",")}}
	}
	path := listPath(adsBasePath(teamID, projectID)+"/campaigns", limit, cursor, extra)
	var out Page[AdCampaign]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) GetAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "GET", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID), nil, &out)
}

func (c *Client) SearchTargeting(ctx context.Context, teamID, projectID, dimension, query string, limit int) ([]TargetingOption, error) {
	q := url.Values{"dimension": {dimension}, "q": {query}}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	var out struct {
		Data []TargetingOption `json:"data"`
	}
	path := adsBasePath(teamID, projectID) + "/targeting/search?" + q.Encode()
	return out.Data, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) AdInsights(ctx context.Context, teamID, projectID, campaignID string, daily bool, since, until string) (AdInsightsSeries, error) {
	q := url.Values{}
	if daily {
		q.Set("daily", "true")
	}
	if since != "" {
		q.Set("since", since)
	}
	if until != "" {
		q.Set("until", until)
	}
	path := adsBasePath(teamID, projectID) + "/campaigns/" + url.PathEscape(campaignID) + "/insights"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out AdInsightsSeries
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) UpdateAd(ctx context.Context, teamID, projectID, campaignID string, req AdUpdateRequest) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "PATCH", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID), req, &out)
}

func (c *Client) PauseAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID)+"/pause", nil, &out)
}

func (c *Client) ResumeAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID)+"/resume", nil, &out)
}

func (c *Client) DeleteAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "DELETE", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID), nil, &out)
}

func (c *Client) PixelCreate(ctx context.Context, teamID, projectID string) (Pixel, error) {
	var out Pixel
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/pixel", nil, &out)
}

func (c *Client) PixelStatus(ctx context.Context, teamID, projectID string) (PixelStatus, error) {
	var out PixelStatus
	return out, c.do(ctx, "GET", adsBasePath(teamID, projectID)+"/pixel", nil, &out)
}

func (c *Client) PixelList(ctx context.Context, teamID string, limit int, cursor string) (Page[Pixel], error) {
	base := fmt.Sprintf("/v1/teams/%s/ads/pixels", url.PathEscape(teamID))
	var out Page[Pixel]
	return out, c.do(ctx, "GET", listPath(base, limit, cursor, nil), nil, &out)
}

// SendConversions relays a raw Meta CAPI JSON payload and returns the upstream
// status + body verbatim (no JSON decoding, no >=400 -> error conversion) so the
// caller can print Meta's own response.
func (c *Client) SendConversions(ctx context.Context, teamID, projectID string, payload []byte) (int, []byte, error) {
	status, _, data, err := c.doRequest(ctx, "POST",
		adsBasePath(teamID, projectID)+"/meta/conversions",
		rawBody(payload, "application/json"))
	if err != nil {
		return 0, nil, err
	}
	return status, data, nil
}
