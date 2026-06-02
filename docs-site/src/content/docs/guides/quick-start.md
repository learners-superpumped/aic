---
title: Quick start
description: Log in and run your first aic commands.
---

Once `aic` is [installed](/guides/installation/), sign in and start working.

## Log in

```bash
aic login
```

This opens an OIDC browser login. Your AIC account session is stored locally and
refreshed automatically.

## Explore your teams and projects

```bash
aic teams list
aic projects list --team <team-id>
```

## Search for a domain

```bash
aic domains search example.com --team <team-id>
```

## Output formats

Every command accepts `--output` (`-o`) — `table` (default), `json`, or `yaml` —
which makes `aic` easy to script for both humans and AI agents:

```bash
aic teams list --output json
```

## Getting help

```bash
aic --help
aic <command> --help
```

The full, always-current command reference lives under
[Command reference](/reference/aic/).
