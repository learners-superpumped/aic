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

// AuthStatus is the account snapshot shown by `aic auth status`: who you are,
// your current working context, your credit, and the teams you belong to.
type AuthStatus struct {
	UserID           string   `json:"user_id" yaml:"user_id"`
	Email            string   `json:"email,omitempty" yaml:"email,omitempty"`
	APIEndpoint      string   `json:"api_endpoint" yaml:"api_endpoint"`
	DefaultTeamID    string   `json:"default_team_id,omitempty" yaml:"default_team_id,omitempty"`
	DefaultTeam      *Team    `json:"default_team,omitempty" yaml:"default_team,omitempty"`
	DefaultProjectID string   `json:"default_project_id,omitempty" yaml:"default_project_id,omitempty"`
	DefaultProject   *Project `json:"default_project,omitempty" yaml:"default_project,omitempty"`
	BalanceUSD       float64  `json:"balance_usd" yaml:"balance_usd"`
	Teams            []Team   `json:"teams" yaml:"teams"`
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
	ID           string    `json:"id" yaml:"id"`
	AmountNano   int64     `json:"amount_nano" yaml:"amount_nano"`
	Type         string    `json:"type" yaml:"type"`
	ResourceType string    `json:"resource_type,omitempty" yaml:"resource_type,omitempty"`
	Reference    string    `json:"reference" yaml:"reference"`
	CreatedAt    time.Time `json:"created_at" yaml:"created_at"`
}

// UsageBucket is one resource's aggregated usage over a window.
type UsageBucket struct {
	Resource string `json:"resource" yaml:"resource"`
	Entries  int64  `json:"entries" yaml:"entries"`
	SpendUSD string `json:"spend_usd" yaml:"spend_usd"`
}

// UsageSummary is per-resource usage over a date window.
type UsageSummary struct {
	From          string        `json:"from" yaml:"from"`
	To            string        `json:"to" yaml:"to"`
	ByResource    []UsageBucket `json:"by_resource" yaml:"by_resource"`
	TotalSpendUSD string        `json:"total_spend_usd" yaml:"total_spend_usd"`
}

// TopupResult is the response to a top-up request.
type TopupResult struct {
	Status          string `json:"status" yaml:"status"`
	PaymentIntentID string `json:"payment_intent_id" yaml:"payment_intent_id"`
}

// AutoRechargeConfig is a team's auto top-up policy.
type AutoRechargeConfig struct {
	Enabled         bool    `json:"enabled" yaml:"enabled"`
	ThresholdUSD    float64 `json:"threshold_usd" yaml:"threshold_usd"`
	AmountUSD       float64 `json:"amount_usd" yaml:"amount_usd"`
	MonthlyLimitUSD float64 `json:"monthly_limit_usd" yaml:"monthly_limit_usd"`
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

type Invite struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	InvitedBy string    `json:"invited_by"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InvitePreview struct {
	TeamName       string `json:"team_name"`
	Role           string `json:"role"`
	InvitedByEmail string `json:"invited_by_email"`
}

type Member struct {
	UserSub  string `json:"user_sub"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
}

// AdTargeting is the audience-targeting parameters for an ad campaign.
type AdTargeting struct {
	Geo               []string `json:"geo,omitempty"`
	AgeMin            int      `json:"ageMin,omitempty"`
	AgeMax            int      `json:"ageMax,omitempty"`
	Genders           []string `json:"genders,omitempty"`
	Interests         []string `json:"interests,omitempty"`
	CustomAudienceRef *string  `json:"customAudienceRef,omitempty"`
}

// AdCreative is the creative content for an ad campaign.
type AdCreative struct {
	StorageRef     string `json:"storageRef,omitempty"`
	Headline       string `json:"headline,omitempty"`
	Body           string `json:"body,omitempty"`
	CTA            string `json:"cta,omitempty"`
	DestinationURL string `json:"destinationUrl,omitempty"`
}

// AdInsights is one window of campaign performance.
type AdInsights struct {
	DateStart   string  `json:"date_start" yaml:"date_start"`
	DateStop    string  `json:"date_stop" yaml:"date_stop"`
	SpendNano   int64   `json:"spend_nano" yaml:"spend_nano"`
	Impressions int64   `json:"impressions" yaml:"impressions"`
	Clicks      int64   `json:"clicks" yaml:"clicks"`
	Reach       int64   `json:"reach" yaml:"reach"`
	Frequency   float64 `json:"frequency" yaml:"frequency"`
	CTR         float64 `json:"ctr" yaml:"ctr"`
	CPCNano     int64   `json:"cpc_nano" yaml:"cpc_nano"`
	CPMNano     int64   `json:"cpm_nano" yaml:"cpm_nano"`
	Results     int64   `json:"results" yaml:"results"`
	ResultType  string  `json:"result_type,omitempty" yaml:"result_type,omitempty"`
}

// AdInsightsSeries is the cumulative window plus an optional daily series.
type AdInsightsSeries struct {
	Cumulative AdInsights   `json:"cumulative" yaml:"cumulative"`
	Daily      []AdInsights `json:"daily,omitempty" yaml:"daily,omitempty"`
}

// AdUpdateRequest carries in-place edits to a running campaign.
type AdUpdateRequest struct {
	BudgetNano *int64       `json:"budgetNano,omitempty"`
	Targeting  *AdTargeting `json:"targeting,omitempty"`
	Placements []string     `json:"placements,omitempty"`
	EndAt      *time.Time   `json:"endAt,omitempty"`
}

// AdLaunchRequest is the payload for launching an ad campaign.
type AdLaunchRequest struct {
	LaunchToken     string         `json:"launchToken"`
	Objective       string         `json:"objective"`
	BudgetType      string         `json:"budgetType"`
	BudgetNano      int64          `json:"budgetNano"`
	StartAt         time.Time      `json:"startAt"`
	EndAt           *time.Time     `json:"endAt,omitempty"`
	Targeting       AdTargeting    `json:"targeting,omitempty"`
	Placements      []string       `json:"placements,omitempty"`
	ProviderOptions map[string]any `json:"providerOptions,omitempty"`
	Creative        AdCreative     `json:"creative"`
}

// AdSchedule is the start/end window of a campaign.
type AdSchedule struct {
	StartAt time.Time  `json:"start_at" yaml:"start_at"`
	EndAt   *time.Time `json:"end_at,omitempty" yaml:"end_at,omitempty"`
}

// AdCampaign is a launched ad campaign as returned by the API.
type AdCampaign struct {
	ID           string     `json:"id" yaml:"id"`
	TeamID       string     `json:"team_id" yaml:"team_id"`
	ProjectID    string     `json:"project_id" yaml:"project_id"`
	Provider     string     `json:"provider" yaml:"provider"`
	Objective    string     `json:"objective" yaml:"objective"`
	BudgetType   string     `json:"budget_type" yaml:"budget_type"`
	BudgetNano   int64      `json:"budget_nano" yaml:"budget_nano"`
	Schedule     AdSchedule `json:"schedule" yaml:"schedule"`
	Status       string     `json:"status" yaml:"status"`
	StatusReason string     `json:"status_reason,omitempty" yaml:"status_reason,omitempty"`
	ExternalID   string     `json:"external_id,omitempty" yaml:"external_id,omitempty"`
	LaunchToken  string     `json:"launch_token" yaml:"launch_token"`
	// SpentNano is actual spend drawn from Meta; ReservedNano is still encumbered.
	SpentNano    int64     `json:"spent_nano" yaml:"spent_nano"`
	ReservedNano int64     `json:"reserved_nano" yaml:"reserved_nano"`
	CreatedAt    time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" yaml:"updated_at"`
}
