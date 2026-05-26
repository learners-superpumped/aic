# aicompany-platform Monorepo Migration — Design Spec

- **Date:** 2026-05-26
- **Status:** Approved design — pending implementation plan
- **Type:** One-time structural migration (no feature work)

## 1. Purpose

Consolidate the two existing, independently-pushed repositories into a single
**multi-module Go monorepo**, preserving git history, then retire the old
repositories (local directories and GitHub remotes).

- `aicompany-cli` (local `~/Developments/aic`) — the `aic` CLI
- `aicompany-auth` (local `~/Developments/aic-auth`) — the central OIDC IdP (Zitadel config + `pkg/oidcauth`)

Target: `aicompany-platform` (GitHub `learners-superpumped/aicompany-platform`, local `~/Developments/aicompany-platform`).

### Why monorepo, why multi-module
- A small, single team building tightly-related Go services that **share code** (`pkg/oidcauth` is consumed by the future backend). A monorepo removes private-module version pinning and `replace` friction, and the current Go module paths (`github.com/learners-superpumped/...`) don't even match the GitHub org (`learners-superpumped`) — a single workspace sidesteps cross-repo `go get` entirely.
- **Multi-module + `go.work`** is the standard Go monorepo shape for independently-releasable components (CLI / API / shared lib) and is what makes **history-preserving subdirectory import** clean (each repo moves into one subdir, no import-path rewriting).

## 2. Target Structure

```
aicompany-platform/                 (one git repo)
├── go.work                         # use ./cli ./auth   (./api added in the backend slice)
├── README.md
├── cli/                            # = former aicompany-cli, moved wholesale (own go.mod)
│   ├── go.mod                      #   module github.com/learners-superpumped/aic   (unchanged)
│   ├── main.go, cmd/, internal/{api,app,auth,config,output}/
│   └── docs/superpowers/...
└── auth/                           # = former aicompany-auth, moved wholesale (own go.mod)
    ├── go.mod                      #   module github.com/learners-superpumped/aic-auth (unchanged)
    ├── pkg/oidcauth/
    ├── deploy/{local,terraform}/
    └── docs/superpowers/...
```

- `go.work` ties the modules for local builds/tests. The future `api` module imports `pkg/oidcauth` by its existing module path; `go.work` resolves it locally (no `go get`, no path rewrite).
- **Module paths are left unchanged.** Renaming them to `github.com/learners-superpumped/aicompany-platform/...` would require rewriting every import and break clean history preservation; inside the workspace the existing paths work. Renaming is an optional, separate later cleanup with no functional impact.

## 3. Migration Method (history-preserving)

For each source repo, rewrite its history so all paths are nested under its target
subdirectory, then merge into the monorepo with unrelated histories allowed:

1. `git init` the new `aicompany-platform`; add an initial root commit (README + `.gitignore`).
2. For `aicompany-cli` → `cli/`:
   - Work on a **fresh clone** (never the original): `git clone <local aic> /tmp/mig-cli`.
   - `cd /tmp/mig-cli && git filter-repo --to-subdirectory-filter cli`.
   - In the monorepo: `git remote add cli-src /tmp/mig-cli && git fetch cli-src && git merge --allow-unrelated-histories cli-src/main` (resolve the branch name from the clone), then `git remote remove cli-src`.
3. Repeat for `aicompany-auth` → `auth/` (clone to `/tmp/mig-auth`, `--to-subdirectory-filter auth`, merge).
4. Add `go.work` (`go 1.26` + `use ./cli ./auth`), root `README.md`, root `.gitignore`.

This preserves the full commit history of both repos under `cli/` and `auth/`.

## 4. Verification (before any deletion)

All must pass in the new monorepo:
- `go build ./...` and `go test ./...` (run with the workspace active — from repo root with `go.work` present).
- `cd auth/deploy/terraform && terraform init -backend=false && terraform validate`.
- `git log --oneline -- cli/ | tail` and `... -- auth/` show the imported history (sanity that history survived).

## 5. Publish, then Retire (destructive — gated)

1. Create the GitHub repo and push: `gh repo create learners-superpumped/aicompany-platform --private --source=. --remote=origin --push`.
2. Confirm the push (browse / `gh repo view`).
3. **Only after the new repo is verified present**, with one explicit final confirmation:
   - Delete old local dirs: `~/Developments/aic`, `~/Developments/aic-auth`.
   - Delete old GitHub remotes: `gh repo delete learners-superpumped/aicompany-cli --yes`, `gh repo delete learners-superpumped/aicompany-auth --yes`.

GitHub repo deletion is irreversible; it runs last, after the monorepo is confirmed pushed.

## 6. Follow-ups (out of scope here)

- Update the memory file paths (`aic-platform-progress.md`) to the new monorepo layout.
- Backend resource server (slice 1) is a **separate** sub-project, built afterward as the `./api` module inside this monorepo.
- Optional later: rename module paths to `github.com/learners-superpumped/aicompany-platform/{cli,auth,api}` (cosmetic; no functional need).

## 7. Risks & Mitigations

- **Data loss on delete** → new repo is created, pushed, and verified before any deletion; final confirmation required.
- **History corruption** → all filter-repo work runs on throwaway clones in `/tmp`, never the originals; originals remain intact until the explicit delete step.
- **`go.work` not resolving** → verified by `go build ./...` from the workspace root before publishing.
- **Open IDE files / shell cwd** referencing old paths break after the move — expected; re-open from the new path.
- **`gh repo delete` scope** → requires the `delete_repo` token scope; if absent, surface it and have the user delete via the GitHub UI rather than failing silently.
