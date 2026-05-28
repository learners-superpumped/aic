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

// Domain is a domain in a project.
type Domain struct {
	Name         string    `json:"name" yaml:"name"`
	Source       string    `json:"source,omitempty" yaml:"source,omitempty"`
	Status       string    `json:"status" yaml:"status"`
	AutoRenew    bool      `json:"auto_renew" yaml:"auto_renew"`
	RegisteredAt time.Time `json:"registered_at,omitempty" yaml:"registered_at,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	VerifiedAt   time.Time `json:"verified_at,omitempty" yaml:"verified_at,omitempty"`
	LastVerifyAt time.Time `json:"last_verify_at,omitempty" yaml:"last_verify_at,omitempty"`
	HostedZoneID string    `json:"hosted_zone_id,omitempty" yaml:"hosted_zone_id,omitempty"`
	Nameservers  []string  `json:"nameservers,omitempty" yaml:"nameservers,omitempty"`
}

// ConnectDomainResponse is returned by POST .../domains/connect.
type ConnectDomainResponse struct {
	Domain Domain `json:"domain" yaml:"domain"`
}

// VerifyDomainResponse is returned by POST .../domains/{name}/verify.
type VerifyDomainResponse struct {
	Domain    Domain    `json:"domain" yaml:"domain"`
	Observed  []string  `json:"observed" yaml:"observed"`
	Expected  []string  `json:"expected" yaml:"expected"`
	CheckedAt time.Time `json:"checked_at" yaml:"checked_at"`
}

// DomainContactInput is the create/update request payload — writable fields only.
// (Response timestamps are not in the input shape; the server's strict decoder
// rejects them.)
type DomainContactInput struct {
	Name         string `json:"name"`
	IsDefault    bool   `json:"is_default"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state,omitempty"`
	Zip          string `json:"zip"`
	Country      string `json:"country"`
}

// DomainContact is a per-team WHOIS contact profile.
type DomainContact struct {
	Name         string    `json:"name" yaml:"name"`
	IsDefault    bool      `json:"is_default" yaml:"is_default"`
	FirstName    string    `json:"first_name" yaml:"first_name"`
	LastName     string    `json:"last_name" yaml:"last_name"`
	Organization string    `json:"organization,omitempty" yaml:"organization,omitempty"`
	Email        string    `json:"email" yaml:"email"`
	Phone        string    `json:"phone" yaml:"phone"`
	AddressLine1 string    `json:"address_line1" yaml:"address_line1"`
	AddressLine2 string    `json:"address_line2,omitempty" yaml:"address_line2,omitempty"`
	City         string    `json:"city" yaml:"city"`
	State        string    `json:"state,omitempty" yaml:"state,omitempty"`
	Zip          string    `json:"zip" yaml:"zip"`
	Country      string    `json:"country" yaml:"country"`
	CreatedAt    time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" yaml:"updated_at"`
}

// DomainSearchResult is one availability/pricing row.
type DomainSearchResult struct {
	Domain    string  `json:"domain" yaml:"domain"`
	Available bool    `json:"available" yaml:"available"`
	PriceUSD  float64 `json:"price_usd" yaml:"price_usd"`
	Currency  string  `json:"currency" yaml:"currency"`
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

// CreditBalance is a team's wallet balance.
type CreditBalance struct {
	BalanceNano int64   `json:"balance_nano" yaml:"balance_nano"`
	BalanceUSD  float64 `json:"balance_usd" yaml:"balance_usd"`
}

// LedgerEntry is one credit ledger row.
type LedgerEntry struct {
	ID         string    `json:"id" yaml:"id"`
	AmountNano int64     `json:"amount_nano" yaml:"amount_nano"`
	Type       string    `json:"type" yaml:"type"`
	Reference  string    `json:"reference" yaml:"reference"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
}

// TopupResult is the response to a top-up request.
type TopupResult struct {
	Status          string `json:"status" yaml:"status"`
	PaymentIntentID string `json:"payment_intent_id" yaml:"payment_intent_id"`
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
