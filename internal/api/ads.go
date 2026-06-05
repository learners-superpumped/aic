package api

import (
	"context"
	"fmt"
	"net/url"
)

func adsBasePath(teamID, projectID string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/ads",
		url.PathEscape(teamID), url.PathEscape(projectID))
}

func (c *Client) LaunchAd(ctx context.Context, teamID, projectID string, req AdLaunchRequest) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "POST", adsBasePath(teamID, projectID)+"/campaigns", req, &out)
}

func (c *Client) ListAds(ctx context.Context, teamID, projectID string) ([]AdCampaign, error) {
	var out []AdCampaign
	return out, c.do(ctx, "GET", adsBasePath(teamID, projectID)+"/campaigns", nil, &out)
}

func (c *Client) GetAd(ctx context.Context, teamID, projectID, campaignID string) (AdCampaign, error) {
	var out AdCampaign
	return out, c.do(ctx, "GET", adsBasePath(teamID, projectID)+"/campaigns/"+url.PathEscape(campaignID), nil, &out)
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
