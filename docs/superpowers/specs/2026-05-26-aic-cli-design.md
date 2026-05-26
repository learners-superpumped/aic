# aic CLI — Design Spec

- **Date:** 2026-05-26
- **Status:** Approved design — pending implementation plan
- **Scope of this spec:** CLI command interface + the backend API contract the CLI depends on. Backend implementation is a separate, follow-up action.

> Note on language: this document and all project discussion are in Korean, but **every user-facing string emitted by the CLI — help, usage, flag descriptions, error messages, and command output — MUST be written in English.**

---

## 1. Overview

`aic` is a Go CLI that lets users (and AI agents) provision and pay for resources on our service. After downloading the CLI a user can:

1. Authenticate and obtain credentials from our backend (browser-based login).
2. Register a payment method (card) via a hosted browser flow.
3. Create **projects**, buy **domains**, create email **inboxes** on those domains, and **send/list messages**.

The CLI talks **only to our backend API** (single backend). The backend internally integrates external services (Stripe for payments, AWS SES for email, a domain registrar for domains). The CLI is unaware of these external services.

### Conceptual references
- **AWS CLI** — credentials/profile UX (`~/.aws`-style config), `noun verb` command structure, `-o`/`--output` formats.
- **AgentMail** — concept reference only for "email inboxes for agents." We do **not** use AgentMail; inboxes are implemented on **AWS SES** in our backend.

---

## 2. Resource Model

```
Account            (login, billing/card)
└── Project        (many)
    └── Domain     (many; purchased through the project)
        └── Inbox  (many; created only on a purchased & verified domain)
            └── Message (send / list / show)
```

Key relationship: **an inbox can only be created on a domain the project owns.** Buying a domain triggers backend-side SES domain verification/DKIM setup; once verified, inboxes (e.g. `agent@example.com`) can be created on it. No domain → no inbox.

---

## 3. Tech Stack

- **Language:** Go
- **CLI framework:** `spf13/cobra` (de-facto standard; used by kubectl, gh, hugo, helm). Supports nested commands, auto help, shell completion.
- **Config:** `spf13/viper` for layered configuration (file → env → flag).
- **Rationale:** the `noun verb` nested structure plus multi-source config is exactly Cobra+Viper's sweet spot, and yields a UX familiar to gh/kubectl users.

---

## 4. Command Structure

**Style:** resource-action (`noun verb`), **max depth 2**. Project membership is handled via **context** (a default project or the `--project` flag), never as a third path segment. This avoids verbose 3-level trees like `aic projects <id> domains list`.

```
aic login                              # OAuth device flow (browser)
aic logout
aic whoami
aic configure                          # interactive setup (profile, default project, output)

aic projects list                      # alias: project, proj ; list alias: ls
aic projects create <name>
aic projects delete <id>               # alias: rm
aic projects use <id>                  # set default project in config
aic projects show <id>

aic domains search <query>             # alias: domain ; availability + pricing
aic domains buy <domain>               # register/purchase (requires card on file)
aic domains list
aic domains show <domain>

aic inboxes create <address>           # alias: inbox ; address must be on an owned domain
aic inboxes list
aic inboxes delete <address>           # alias: rm
aic inboxes show <address>

aic messages send                      # alias: msg, mail ; flags: --from --to --subject --body/--body-file
aic messages list                      # for a given inbox (--inbox)
aic messages show <id>                 # alias: read

aic billing add-card                   # hosted browser flow (Stripe)
aic billing cards                      # list registered cards
aic billing status                     # billing/account status
```

Frequently-used auth commands (`login`, `logout`, `whoami`) live at the top level as single-level commands by design.

---

## 5. Cross-Cutting Concerns (Modularization)

### 5.1 Persistent flags (defined once on root, inherited everywhere)
```go
root.PersistentFlags().StringP("project", "p", "", "target project (overrides the default project)")
root.PersistentFlags().StringP("output",  "o", "table", "output format: table|json|yaml")
root.PersistentFlags().String("profile", "default", "credentials profile to use")
```

### 5.2 Context resolution as middleware (`PersistentPreRunE`)
Runs before every command; resolves project as `--project flag > config default_project > error`, builds the shared app context, injects it:
```go
root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load(profile)            // load ~/.aic
    if err != nil { return err }
    proj, err := resolveProject(projectFlag, cfg) // flag > default_project
    if err != nil { return err }
    cmd.SetContext(app.New(cfg, proj, output))    // authed client + project + renderer
    return nil
}
```
Each command reads everything it needs with one line: `a := app.FromContext(cmd)`.

### 5.3 Commands that require a project vs not
- **No project required:** `login`, `logout`, `whoami`, `configure`, `projects list`, `projects create`, `billing *`.
- **Project required:** `domains *`, `inboxes *`, `messages *`, `projects show/delete/use`. These error clearly (English) when no project is resolved, suggesting `aic projects use <id>` or `--project`.

---

## 6. Package Layout

```
aic/
├── main.go
├── cmd/                  # thin command definitions: flag parsing + validation only, no business logic
│   ├── root.go          #   persistent flags + PersistentPreRunE middleware
│   ├── auth.go          #   login / logout / whoami / configure
│   ├── projects.go
│   ├── domains.go
│   ├── inboxes.go
│   ├── messages.go
│   └── billing.go
└── internal/
    ├── app/             # shared context: { client, project, output renderer }; FromContext helper
    ├── api/             # backend API client — single entry point for ALL HTTP calls; structured errors
    ├── config/          # ~/.aic credentials + config load/save (AWS-style INI), file perms 0600
    ├── auth/            # device flow start/poll, token refresh
    └── output/          # table / json / yaml renderers shared by every command
```

**Principle:** `cmd/` stays thin (flags + validation); real work lives in `internal/`. One place per concern → fix once, applies everywhere; each package is independently testable.

---

## 7. Configuration & Credential Storage

AWS-style, INI-formatted, under `~/.aic/` (override via `AIC_CONFIG_DIR` env).

`~/.aic/credentials` (perms `0600`):
```ini
[default]
access_token  = ...
refresh_token = ...
id_token      = ...
expires_at    = 2026-05-26T12:00:00Z
```

`~/.aic/config`:
```ini
[default]
default_project = proj_abc123
output          = table
api_endpoint    = https://api.aic.example.com
issuer          = https://auth.example.com
client_id       = <oidc-public-client-id>
```

The OIDC `issuer` and `client_id` are set via `aic configure --issuer <url> --client-id <id>` (values come from the `aic-auth` Terraform outputs `issuer_url` / `aic_cli_client_id`). The `id_token` is stored for client-side `whoami`.

Multiple profiles supported via section names; selected with `--profile`. Credentials-file values take precedence over config for overlapping keys. Token refresh happens transparently in `auth/`; on unrecoverable auth failure the CLI prints an English message instructing the user to run `aic login`.

---

## 8. Browser-Delegated Flows

Two browser-delegated flows exist; they no longer share one backend mechanism.

**`aic login` — standard OIDC (against the central IdP, not this backend):**
- Default: loopback **Authorization Code + PKCE** — the CLI opens the browser to the issuer's authorize endpoint, receives the code on a `127.0.0.1` callback, and exchanges it (with the PKCE verifier) for tokens.
- `--headless`: **Device Authorization Grant** — the CLI shows a URL + user code to enter in any browser, then polls the token endpoint.
- On success the CLI writes the access/refresh/ID tokens to `~/.aic/credentials`.

**`aic billing add-card` — custom card-session flow (against this backend):**
1. CLI calls `POST /v1/billing/card-sessions` → `{ session_id, browser_url, poll_token, ... }`.
2. CLI opens the browser to the Stripe-hosted URL (prints it as a headless fallback).
3. CLI polls `GET /v1/billing/card-sessions/{id}` until `completed` (or `expired`/`denied`).
4. Stripe-hosted page captures the card; backend confirms; CLI reports success.

PCI scope stays with the backend/PG (Stripe). No card data ever touches the terminal. The reusable poll loop (`auth.RunFlow`) backs the add-card flow; login uses the OIDC-specific loopback/device flows.

---

## 9. Output & Error Handling

- **Output:** default `table` (human-readable). `-o json` and `-o yaml` for scripting/agents. All rendering goes through `internal/output`; commands never `fmt.Print` results directly.
- **Errors:** `internal/api` returns structured errors (HTTP status + backend error code + message). `cmd/` surfaces them as concise English messages. Validation errors (e.g. inbox address not on an owned domain) are caught before the API call where possible. Non-zero exit codes on failure for scriptability.

---

## 10. Backend API Contract (draft)

The CLI depends on these endpoints. Shapes are indicative; the backend implementation is a follow-up action that will finalize them.

### Auth (delegated to the central OIDC IdP)
Authentication is handled by the central OIDC IdP (see the `aic-auth` spec), not by
this backend. The CLI logs in via standard OIDC against the configured issuer —
loopback Authorization Code + PKCE by default, or the Device Authorization Grant
with `--headless` — stores the resulting tokens in `~/.aic`, and sends a Bearer
access JWT on every API call. Token refresh uses the OIDC `refresh_token` grant at
the issuer's token endpoint (transparently, on a `401`). This backend issues no auth
tokens itself; it validates the Bearer JWT (issuer/audience/expiry via JWKS) using
the shared `oidcauth` middleware and reads `sub`/`email` from the verified token.
`whoami` is served client-side from the stored ID token (no backend call).

### Billing
- `POST /v1/billing/card-sessions` → `{ session_id, browser_url, poll_token, expires_at }` (Stripe-hosted)
- `GET  /v1/billing/card-sessions/{id}` (poll) → `{ status, ... }`
- `GET  /v1/billing/cards` → `[{ card_id, brand, last4, exp_month, exp_year, default }]`
- `GET  /v1/billing/status` → `{ has_payment_method, ... }`

### Projects
- `GET    /v1/projects` → `[{ id, name, created_at }]`
- `POST   /v1/projects` `{ name }` → `{ id, name, created_at }`
- `GET    /v1/projects/{id}` → `{ id, name, ... }`
- `DELETE /v1/projects/{id}` → `204`

### Domains (scoped to a project)
- `GET    /v1/projects/{pid}/domains/search?q=` → `[{ domain, available, price, currency }]`
- `POST   /v1/projects/{pid}/domains` `{ domain }` → `{ domain, status, verification: {...} }`
- `GET    /v1/projects/{pid}/domains` → `[{ domain, status, ... }]`
- `GET    /v1/projects/{pid}/domains/{domain}` → `{ domain, status, dkim/spf/dmarc, ... }`

### Inboxes (scoped to a project; address must be on an owned, verified domain)
- `POST   /v1/projects/{pid}/inboxes` `{ address }` → `{ address, status, ... }`
- `GET    /v1/projects/{pid}/inboxes` → `[{ address, status, ... }]`
- `GET    /v1/projects/{pid}/inboxes/{address}` → `{ address, status, ... }`
- `DELETE /v1/projects/{pid}/inboxes/{address}` → `204`

### Messages (scoped to an inbox)
- `POST /v1/projects/{pid}/inboxes/{address}/messages` `{ to, subject, body }` → `{ message_id, status }`
- `GET  /v1/projects/{pid}/inboxes/{address}/messages` → `[{ message_id, from, to, subject, snippet, received_at }]`
- `GET  /v1/projects/{pid}/inboxes/{address}/messages/{id}` → `{ message_id, headers, body, ... }`

All requests carry `Authorization: Bearer <access_token>`.

---

## 11. V1 Scope

Full set: auth (`login`/`logout`/`whoami`/`configure`) + billing (`add-card`/`cards`/`status`) + projects CRUD + domains (search/buy/list/show) + inboxes (create/list/delete/show) + messages (send/list/show).

### Out of scope (later)
- Public/shared-domain inboxes (V1 requires a purchased domain).
- Team/role management, multi-user orgs.
- Webhooks / push delivery of incoming mail.
- Non-Stripe payment providers.

---

## 12. Testing Strategy

- Each `internal/` package independently unit-tested.
- `internal/api` tested against an `httptest` mock backend (covers structured-error mapping and token refresh).
- `internal/config` tested for round-trip load/save, profile selection, and `0600` perms.
- `cmd/` tested by invoking Cobra commands with a faked `app` context (table/json output assertions); no real network.
- Browser-delegated flows tested with a stubbed browser-open + mocked poll responses (pending → completed / expired).
