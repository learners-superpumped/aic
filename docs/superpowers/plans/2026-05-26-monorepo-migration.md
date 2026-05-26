# aicompany-platform Monorepo Migration — Implementation Plan

> **For agentic workers:** This is a stateful, partly-irreversible migration (it deletes GitHub repos at the end). Execute it **inline with checkpoints** (superpowers:executing-plans), NOT via parallel subagents. Steps use checkbox (`- [ ]`) syntax. The destructive Task 6 has a HARD GATE requiring explicit user confirmation.

**Goal:** Consolidate `aicompany-cli` (~/Developments/aic) and `aicompany-auth` (~/Developments/aic-auth) into one history-preserving multi-module Go monorepo `aicompany-platform`, publish it, then retire the old repos.

**Architecture:** Multi-module monorepo tied by `go.work`. Each source repo is imported wholesale into a subdirectory (`cli/`, `auth/`) via `git filter-repo --to-subdirectory-filter`, preserving full history. Module paths are left unchanged (the workspace resolves them locally; no import rewriting). Destructive cleanup runs only after the new repo is pushed and verified.

**Tech Stack:** git, git-filter-repo (installed), Go 1.26 workspaces (`go.work`), gh CLI, Terraform (validate only).

**Reference spec:** `docs/superpowers/specs/2026-05-26-monorepo-migration-design.md`

**Targets:** repo `learners-superpumped/aicompany-platform`, local `~/Developments/aicompany-platform`.

---

## Pre-flight (verify before starting)

- [ ] **Step 1: Confirm sources are clean and pushed**

Run:
```bash
for d in aic aic-auth; do
  echo "== $d =="; git -C ~/Developments/$d status --short; \
  echo "branch: $(git -C ~/Developments/$d branch --show-current)"; \
  git -C ~/Developments/$d log --oneline @{u}..HEAD 2>/dev/null | head
done
git-filter-repo --version
```
Expected: both repos clean working tree, on `main`, **no unpushed commits** (so the originals' remotes match local). `git-filter-repo --version` prints a version. If there are unpushed commits, push them first (the originals stay the safety net until Task 6).

---

## Task 1: Scaffold the monorepo

**Files:**
- Create: `~/Developments/aicompany-platform/README.md`
- Create: `~/Developments/aicompany-platform/.gitignore`

- [ ] **Step 1: Init the repo with an initial root commit**

Run:
```bash
mkdir -p ~/Developments/aicompany-platform
cd ~/Developments/aicompany-platform
git init -q
```

- [ ] **Step 2: Write README.md**

Create `~/Developments/aicompany-platform/README.md`:
```markdown
# aicompany-platform

Monorepo for the aic platform. Multi-module Go workspace (`go.work`).

- `cli/` — the `aic` CLI (OIDC login, project/domain/inbox provisioning).
- `auth/` — central OIDC IdP: Zitadel deployment + `pkg/oidcauth` (shared JWT verifier/middleware).
- `api/` — backend resource server (added in a later slice).

## Development

```
go build ./...   # builds all workspace modules
go test ./...    # tests all workspace modules
```

Each top-level directory is its own Go module; `go.work` ties them together for
local development. History of `cli/` and `auth/` was imported (preserved) from
the former `aicompany-cli` and `aicompany-auth` repositories.
```

- [ ] **Step 3: Write a root .gitignore**

Create `~/Developments/aicompany-platform/.gitignore`:
```gitignore
# OS
.DS_Store

# Go build output at repo root (per-module ignores live in each module)
/aic
*.test
*.out
```

- [ ] **Step 4: Commit the scaffold**

Run:
```bash
cd ~/Developments/aicompany-platform
git add README.md .gitignore
git commit -q -m "chore: initialize aicompany-platform monorepo"
git log --oneline | cat
```
Expected: one commit `chore: initialize aicompany-platform monorepo`. Note the default branch name (`git branch --show-current` → likely `main` or `master`; if `master`, rename: `git branch -m main`).

---

## Task 2: Import `aicompany-cli` → `cli/` (history preserved)

- [ ] **Step 1: Fresh clone of the CLI repo (never touch the original)**

Run:
```bash
rm -rf /tmp/mig-cli
git clone ~/Developments/aic /tmp/mig-cli
cd /tmp/mig-cli && git branch --show-current
```
Expected: a clone at `/tmp/mig-cli`, branch `main`.

- [ ] **Step 2: Rewrite history under `cli/`**

Run:
```bash
cd /tmp/mig-cli
git-filter-repo --to-subdirectory-filter cli --force
git log --oneline | head -3 | cat
ls
```
Expected: history rewritten so every tree is under `cli/`; `ls` shows only `cli/`. (`git-filter-repo` removes the `origin` remote — expected.)

- [ ] **Step 3: Merge into the monorepo with unrelated histories**

Run:
```bash
cd ~/Developments/aicompany-platform
git remote add cli-src /tmp/mig-cli
git fetch cli-src
git merge --allow-unrelated-histories -m "Import aicompany-cli into cli/ (history preserved)" cli-src/main
git remote remove cli-src
ls cli/
```
Expected: clean merge (no conflicts — `cli/` paths don't overlap the root README/.gitignore). `ls cli/` shows `main.go cmd internal go.mod docs ...`.

- [ ] **Step 4: Verify CLI history survived**

Run:
```bash
cd ~/Developments/aicompany-platform
git log --oneline -- cli/ | tail -3 | cat
git log --oneline -- cli/ | wc -l
```
Expected: the CLI's original commits appear (e.g. the scaffold/feat commits), count > 1.

---

## Task 3: Import `aicompany-auth` → `auth/` (history preserved)

- [ ] **Step 1: Fresh clone of the auth repo**

Run:
```bash
rm -rf /tmp/mig-auth
git clone ~/Developments/aic-auth /tmp/mig-auth
cd /tmp/mig-auth && git branch --show-current
```
Expected: clone at `/tmp/mig-auth`, branch `main`.

- [ ] **Step 2: Rewrite history under `auth/`**

Run:
```bash
cd /tmp/mig-auth
git-filter-repo --to-subdirectory-filter auth --force
ls
```
Expected: `ls` shows only `auth/`.

- [ ] **Step 3: Merge into the monorepo**

Run:
```bash
cd ~/Developments/aicompany-platform
git remote add auth-src /tmp/mig-auth
git fetch auth-src
git merge --allow-unrelated-histories -m "Import aicompany-auth into auth/ (history preserved)" auth-src/main
git remote remove auth-src
ls auth/
```
Expected: clean merge; `ls auth/` shows `go.mod pkg deploy docs ...`.

- [ ] **Step 4: Verify auth history survived**

Run:
```bash
cd ~/Developments/aicompany-platform
git log --oneline -- auth/ | wc -l
git log --oneline -- auth/pkg/oidcauth/ | tail -2 | cat
```
Expected: count > 1; oidcauth commits present.

---

## Task 4: Add the Go workspace and verify the build

**Files:**
- Create: `~/Developments/aicompany-platform/go.work`

- [ ] **Step 1: Create go.work**

Run:
```bash
cd ~/Developments/aicompany-platform
cat > go.work <<'EOF'
go 1.26.3

use (
	./cli
	./auth
)
EOF
```

- [ ] **Step 2: Build the whole workspace**

Run:
```bash
cd ~/Developments/aicompany-platform
go build ./...
```
Expected: builds with no error (workspace mode resolves both modules). If `go build ./...` from the workspace root does not traverse modules on the installed Go version, run `go build ./cli/... ./auth/...` instead and note it.

- [ ] **Step 3: Test the whole workspace**

Run:
```bash
cd ~/Developments/aicompany-platform
go test ./... 2>&1 | tail -12
```
Expected: all packages pass (`cli` packages + `auth/pkg/oidcauth`). If root `./...` doesn't traverse, use `go test ./cli/... ./auth/...`.

- [ ] **Step 4: Validate Terraform still works in its new location**

Run:
```bash
cd ~/Developments/aicompany-platform/auth/deploy/terraform
terraform init -backend=false >/dev/null 2>&1 && terraform validate
```
Expected: `Success! The configuration is valid.`

- [ ] **Step 5: gofmt/vet sanity + commit the workspace**

Run:
```bash
cd ~/Developments/aicompany-platform
gofmt -l cli/ auth/ | grep -v '_test.go' || true   # informational
go vet ./cli/... ./auth/... 2>&1 | tail -5
git add go.work
git commit -q -m "chore: add go.work tying cli and auth modules"
git log --oneline | head -5 | cat
```
Expected: vet clean; commit created. The log shows the scaffold + two import merges + go.work commit, with the imported histories beneath the merges.

---

## Task 5: Publish to GitHub and verify

- [ ] **Step 1: Create the private repo and push**

Run:
```bash
cd ~/Developments/aicompany-platform
gh repo create learners-superpumped/aicompany-platform --private --source=. --remote=origin --push
```
Expected: prints the new repo URL; pushes the current branch and sets upstream.

- [ ] **Step 2: Verify the push landed**

Run:
```bash
gh repo view learners-superpumped/aicompany-platform --json name,visibility,defaultBranchRef -q '.name,.visibility,.defaultBranchRef.name'
git -C ~/Developments/aicompany-platform log --oneline origin/$(git -C ~/Developments/aicompany-platform branch --show-current) | head -3 | cat
```
Expected: repo name/visibility shown; remote log matches local HEAD. **Do not proceed to Task 6 unless this confirms the monorepo is on GitHub.**

---

## Task 6: Retire the old repos (DESTRUCTIVE — HARD GATE)

> **HARD GATE:** Do NOT run any step in this task until Task 5 Step 2 confirmed the monorepo is pushed, AND the user has given explicit final confirmation to delete. Pause and ask: "Monorepo is pushed and verified at <URL>. Confirm deleting old local dirs (~/Developments/aic, ~/Developments/aic-auth) and GitHub repos (aicompany-cli, aicompany-auth)? This is irreversible." Wait for an explicit yes.

- [ ] **Step 1: Confirm gh has delete scope**

Run:
```bash
gh auth status 2>&1 | grep -i 'scopes' || echo "scopes line not shown"
```
If `delete_repo` is missing, run `gh auth refresh -h github.com -s delete_repo` (interactive — the user runs it), or delete via the GitHub UI. Do not force.

- [ ] **Step 2: Delete the old GitHub remotes**

Run:
```bash
gh repo delete learners-superpumped/aicompany-cli --yes
gh repo delete learners-superpumped/aicompany-auth --yes
```
Expected: each prints deletion confirmation.

- [ ] **Step 3: Delete the old local directories**

Run:
```bash
rm -rf ~/Developments/aic ~/Developments/aic-auth
ls ~/Developments | grep -E '^aic' || true
```
Expected: only `aicompany-platform` remains among `aic*` (the old two are gone).

- [ ] **Step 4: Update the memory file to the new paths**

Edit `/Users/ghyeok/.claude/projects/-Users-ghyeok-Developments-aic/memory/aic-platform-progress.md`:
- Replace repo/path references: `~/Developments/aic` → `~/Developments/aicompany-platform/cli`, `~/Developments/aic-auth` → `~/Developments/aicompany-platform/auth`.
- Replace the two-remotes line with: single repo `learners-superpumped/aicompany-platform` (monorepo, multi-module go.work; `cli/`, `auth/`, future `api/`).
- Note the migration was done 2026-05-26 with history preserved; old repos `aicompany-cli`/`aicompany-auth` deleted.

(No commit needed — memory files are outside git.)

---

## Self-Review Notes

- **Spec coverage:** multi-module go.work (Task 4), history-preserving subdir import via filter-repo on throwaway clones (Tasks 2-3), module paths unchanged / no import rewriting (no task touches imports — by design), verification before deletion (Task 4 + Task 5 Step 2), publish-then-retire with hard gate (Tasks 5-6), memory path update (Task 6 Step 4). All spec sections map to a task.
- **Safety:** originals are never modified (filter-repo runs on `/tmp` clones); deletion is gated behind a verified push + explicit user confirmation; gh delete-scope is checked, not forced.
- **No placeholders:** every step has concrete commands + expected output.
- **Consistency:** subdir names `cli`/`auth` and the `cli-src`/`auth-src` temporary remotes are used consistently; `go.work` lists `./cli ./auth`; branch assumed `main` with an explicit check/rename in Task 1.
- **Branch-name robustness:** Task 1 Step 4 checks/renames the monorepo branch to `main`; Tasks 2-3 verified source branch is `main` in pre-flight/clone steps.
