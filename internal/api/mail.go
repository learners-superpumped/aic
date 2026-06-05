package api

import (
	"context"
	"fmt"
	"mime"
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
	ID     string    `json:"id"`
	From   string    `json:"from"`
	To     []string  `json:"to"`
	SentAt time.Time `json:"sent_at"`
}

type MailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	DataBase64  string `json:"data_base64"`
}

type SendMessageRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	ReplyTo []string `json:"reply_to,omitempty"`
	// ReplyToMessageID, when set, asks the server to build the In-Reply-To /
	// References threading headers from that stored message.
	ReplyToMessageID string           `json:"reply_to_message_id,omitempty"`
	Subject          string           `json:"subject"`
	Text             string           `json:"text,omitempty"`
	HTML             string           `json:"html,omitempty"`
	Attachments      []MailAttachment `json:"attachments,omitempty"`
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

func (c *Client) ListMailDomains(ctx context.Context, teamID, projectID string, limit int, cursor string) (Page[MailIdentity], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := mailBasePath(teamID, projectID) + "/domains"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out Page[MailIdentity]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) DisableMailDomain(ctx context.Context, teamID, projectID, name string) error {
	return c.do(ctx, "DELETE", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(name), nil, nil)
}

func (c *Client) CreateMailInbox(ctx context.Context, teamID, projectID, domain, localPart, displayName string) (*MailInbox, error) {
	var out MailInbox
	body := map[string]string{"local_part": localPart, "display_name": displayName}
	return &out, c.do(ctx, "POST", mailBasePath(teamID, projectID)+"/domains/"+url.PathEscape(domain)+"/inboxes", body, &out)
}

func (c *Client) ListMailInboxes(ctx context.Context, teamID, projectID, domain string, limit int, cursor string) (Page[MailInbox], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := mailBasePath(teamID, projectID) + "/domains/" + url.PathEscape(domain) + "/inboxes"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out Page[MailInbox]
	return out, c.do(ctx, "GET", path, nil, &out)
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

type MailMessage struct {
	ID              string `json:"id"`
	InboxID         string `json:"inbox_id"`
	Direction       string `json:"direction"`
	From            string `json:"from"`
	Subject         string `json:"subject"`
	Snippet         string `json:"snippet"`
	Status          string `json:"status"`
	ThreadID        string `json:"thread_id"`
	ParentMessageID string `json:"parent_message_id,omitempty"`
	InReplyTo       string `json:"in_reply_to,omitempty"`
	SentAt          string `json:"sent_at,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type MailRecipient struct {
	Kind    string `json:"kind"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type MailAttachmentMeta struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type MailMessageDetail struct {
	MailMessage
	TextBody    string               `json:"text_body"`
	HTMLBody    string               `json:"html_body,omitempty"`
	Recipients  []MailRecipient      `json:"recipients"`
	Attachments []MailAttachmentMeta `json:"attachments"`
	RawBase64   string               `json:"raw_base64"`
}

func (c *Client) ListMailMessages(ctx context.Context, teamID, projectID, direction, inbox string, limit int, cursor string) (Page[MailMessage], error) {
	q := url.Values{}
	if direction != "" {
		q.Set("direction", direction)
	}
	if inbox != "" {
		q.Set("inbox", inbox)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	path := mailBasePath(teamID, projectID) + "/messages"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out Page[MailMessage]
	return out, c.do(ctx, "GET", path, nil, &out)
}

func (c *Client) ShowMailMessage(ctx context.Context, teamID, projectID, id string) (*MailMessageDetail, error) {
	var out MailMessageDetail
	return &out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/messages/"+url.PathEscape(id), nil, &out)
}

// MessageThread returns every message in the given message's conversation,
// oldest-first.
func (c *Client) MessageThread(ctx context.Context, teamID, projectID, id string) ([]MailMessage, error) {
	var out []MailMessage
	return out, c.do(ctx, "GET", mailBasePath(teamID, projectID)+"/messages/"+url.PathEscape(id)+"/thread", nil, &out)
}

// GetMailAttachment downloads an attachment's raw bytes and the server-suggested
// filename (from Content-Disposition).
func (c *Client) GetMailAttachment(ctx context.Context, teamID, projectID, messageID, attachmentID string) ([]byte, string, error) {
	path := mailBasePath(teamID, projectID) + "/messages/" + url.PathEscape(messageID) + "/attachments/" + url.PathEscape(attachmentID)
	data, hdr, err := c.doRaw(ctx, "GET", path)
	if err != nil {
		return nil, "", err
	}
	filename := ""
	if _, params, perr := mime.ParseMediaType(hdr.Get("Content-Disposition")); perr == nil {
		filename = params["filename"]
	}
	return data, filename, nil
}
