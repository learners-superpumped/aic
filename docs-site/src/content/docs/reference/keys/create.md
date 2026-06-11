---
title: "aic keys create"
description: "Create an AIC API key (shown once — copy it immediately)"
sidebar:
  label: "create"
---

Create an AIC API key (shown once — copy it immediately)

### Synopsis

Create an AIC API key for server-to-server calls.

Scoped keys carry one or more --scope values (all project-level or all
team-level). Project-level scopes require --project. --full-access mints a
key with the team owner's full authority instead (no --scope/--project).

The raw key is printed once and cannot be recovered; rotate by creating a
new key and revoking the old one.

```
aic keys create [flags]
```

### Options

```
      --expires-in string   key lifetime, e.g. 90d, 12h, 30m (default: no expiry; manage by revoke)
      --full-access         mint a full-access key acting with the team owner's authority (use with care)
  -h, --help                help for create
      --name string         human label for the key
      --scope stringArray   capability scope, repeatable (e.g. --scope storage:read --scope storage:write)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic keys](/reference/keys/keys/)	 - Manage AIC API keys (scoped machine credentials)

