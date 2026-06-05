package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func adsBasePath(teamID, projectID string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/ads",
		url.PathEscape(teamID), url.PathEscape(projectID))
}

func (c *Client) LaunchAd(ctx context.Context, teamID, projectID string, req AdLaunchRequest) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/campaigns", req, &out)
}

func (c *Client) ListAds(ctx context.Context, teamID, projectID string, limit int, cursor string) (Page[AdCampaign], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := adsBasePath(teamID, projectID) + "/campaigns"
	if e := q.Encode(); e != "" {
		path += "?" + e
	}
	var out Page[AdCampaign]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) GetAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "GET", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID), nil, &out)
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
