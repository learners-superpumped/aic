package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

type MailIdentity struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	MailFromDomain string `json:"mail_from_domain"`
	HostedZoneID   string `json:"hosted_zone_id,omitempty"`
	VerifiedAt     string `json:"verified_at,omitempty"`
	LastVerifyAt   string `json:"last_verify_at,omitempty"`
}

type MailDNSRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int32  `json:"ttl"`
}

type EnableMailDomainResponse struct {
	Identity    MailIdentity    `json:"identity"`
	Records     []MailDNSRecord `json:"records"`
	AutoApplied bool            `json:"auto_applied"`
}

type MailInbox struct {
	ID          string `json:"id"`
	Address     string `json:"address"`
	LocalPart   string `json:"local_part"`
	DisplayName string `json:"display_name,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type SendMessageResponse struct {
	SESMessageID string    `json:"ses_message_id"`
	From         string    `json:"from"`
	To           []string  `json:"to"`
	SentAt       time.Time `json:"sent_at"`
}

type MailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	DataBase64  string `json:"data_base64"`
}

type SendMessageRequest struct {
	From        string           `json:"from"`
	To          []string         `json:"to"`
	CC          []string         `json:"cc,omitempty"`
	BCC         []string         `json:"bcc,omitempty"`
	ReplyTo     []string         `json:"reply_to,omitempty"`
	Subject     string           `json:"subject"`
	Text        string           `json:"text,omitempty"`
	HTML        string           `json:"html,omitempty"`
	Attachments []MailAttachment `json:"attachments,omitempty"`
}

func mailBasePath(teamID, projectID string) string {
	return fmt.Sprintf("/v1/teams/%s/projects/%s/mail",
		url.PathEscape(teamID), url.PathEscape(projectID))
}

func (c *Client) EnableMailDomain(ctx context.Context, teamID, projectID, name string) (*EnableMailDomainResponse, error) {
	var out EnableMailDomainResponse
	return &out, c.do(ctx, "POST", mailBasePath(teamID, projectID)+"/domains", map[string]string{"name": name}, &out)
}

func (c *Client) ShowMailDomain(ctx context.Context, teamID, projectID, name string) (*EnableMailDomainResponse, error) {
	var out EnableMailDomainResponse
	return &out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(name), nil, &out)
}

func (c *Client) VerifyMailDomain(ctx context.Context, teamID, projectID, name string) (*EnableMailDomainResponse, error) {
	var out EnableMailDomainResponse
	return &out, c.do(ctx, "POST", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(name)+"/verify", nil, &out)
}

func (c *Client) ListMailDomains(ctx context.Context, teamID, projectID string) ([]MailIdentity, error) {
	var out []MailIdentity
	return out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/domains", nil, &out)
}

func (c *Client) DisableMailDomain(ctx context.Context, teamID, projectID, name string) error {
	return c.do(ctx, "DELETE", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(name), nil, nil)
}

func (c *Client) CreateMailInbox(ctx context.Context, teamID, projectID, domain, localPart, displayName string) (*MailInbox, error) {
	var out MailInbox
	body := map[string]string{"local_part": localPart, "display_name": displayName}
	return &out, c.do(ctx, "POST", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(domain)+"/inboxes", body, &out)
}

func (c *Client) ListMailInboxes(ctx context.Context, teamID, projectID, domain string) ([]MailInbox, error) {
	var out []MailInbox
	return out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(domain)+"/inboxes", nil, &out)
}

func (c *Client) ShowMailInbox(ctx context.Context, teamID, projectID, domain, local string) (*MailInbox, error) {
	var out MailInbox
	return &out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(domain)+"/inboxes/"+url.PathEscape(local), nil, &out)
}

func (c *Client) DeleteMailInbox(ctx context.Context, teamID, projectID, domain, local string) error {
	return c.do(ctx, "DELETE", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(domain)+"/inboxes/"+url.PathEscape(local), nil, nil)
}

func (c *Client) SendMail(ctx context.Context, teamID, projectID string, in SendMessageRequest) (*SendMessageResponse, error) {
	var out SendMessageResponse
	return &out, c.do(ctx, "POST", mailBasePath(teamID, projectID)+"/send", in, &out)
}
