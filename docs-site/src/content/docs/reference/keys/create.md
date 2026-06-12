---
title: "aic keys create"
description: "Create an AIC API key (shown once — copy it immediately)"
sidebar:
  label: "create"
---

Create an AIC API key (shown once — copy it immediately)

### Synopsis

Create an AIC API key for server-to-server calls.

Keys carry one or more --scope values. Project-level scopes require --project.
Use --scope '*' for a full-access key: with --project it is project-owner over
that project's primitives; without --project it is team-owner over the whole
team. The full-access scope cannot be combined with other scopes.

The raw key is printed once and cannot be recovered; rotate by creating a
new key and revoking the old one.

```
aic keys create [flags]
```

### Options

```
      --expires-in string   key lifetime, e.g. 90d, 12h, 30m (default: no expiry; manage by revoke)
  -h, --help                help for create
      --name string         human label for the key
      --scope stringArray   capability scope, repeatable (e.g. --scope storage:read --scope storage:write, or --scope '*')
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

