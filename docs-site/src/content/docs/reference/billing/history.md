---
title: "aic billing history"
description: "Show your AIC credit ledger"
sidebar:
  label: "history"
---

Show your AIC credit ledger

```
aic billing history [flags]
```

### Options

```
      --cursor string   next page's cursor — the 'next_cursor' value from a previous list (shown in -o json output and the table footer)
  -h, --help            help for history
      --limit int       max rows per page (default 50, max 200)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic billing](/reference/billing/billing/)	 - Credits, payment methods, and auto-recharge

