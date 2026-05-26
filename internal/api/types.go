package api

import "time"

// Team is an ownership boundary that contains projects.
type Team struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	Role      string    `json:"role" yaml:"role"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
}

// Project is a provisioning project.
type Project struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
}

// Domain is a purchased domain within a project.
type Domain struct {
	Domain string `json:"domain" yaml:"domain"`
	Status string `json:"status" yaml:"status"`
}

// DomainSearchResult is one availability/pricing row.
type DomainSearchResult struct {
	Domain    string  `json:"domain" yaml:"domain"`
	Available bool    `json:"available" yaml:"available"`
	Price     float64 `json:"price" yaml:"price"`
	Currency  string  `json:"currency" yaml:"currency"`
}

// Inbox is an email inbox on an owned domain.
type Inbox struct {
	Address string `json:"address" yaml:"address"`
	Status  string `json:"status" yaml:"status"`
}

// Message is an email message summary or detail.
type Message struct {
	MessageID  string    `json:"message_id" yaml:"message_id"`
	From       string    `json:"from" yaml:"from"`
	To         string    `json:"to" yaml:"to"`
	Subject    string    `json:"subject" yaml:"subject"`
	Snippet    string    `json:"snippet" yaml:"snippet"`
	Body       string    `json:"body,omitempty" yaml:"body,omitempty"`
	ReceivedAt time.Time `json:"received_at" yaml:"received_at"`
}

// Card is a registered payment method.
type Card struct {
	CardID   string `json:"card_id" yaml:"card_id"`
	Brand    string `json:"brand" yaml:"brand"`
	Last4    string `json:"last4" yaml:"last4"`
	ExpMonth int    `json:"exp_month" yaml:"exp_month"`
	ExpYear  int    `json:"exp_year" yaml:"exp_year"`
	Default  bool   `json:"default" yaml:"default"`
}

// BillingStatus reports whether a payment method exists.
type BillingStatus struct {
	HasPaymentMethod bool `json:"has_payment_method" yaml:"has_payment_method"`
}

// Tokens is an auth token set.
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Session is a browser-delegated session (login or add-card).
type Session struct {
	SessionID  string `json:"session_id"`
	BrowserURL string `json:"browser_url"`
	PollToken  string `json:"poll_token"`
	Status     string `json:"status"` // pending|completed|expired|denied
	*Tokens    `json:",inline"`
}
