---
title: "aic billing auto-recharge enable"
description: "Enable automatic credit top-up"
sidebar:
  label: "enable"
---

Enable automatic credit top-up

```
aic billing auto-recharge enable [flags]
```

### Options

```
      --amount string          amount in USD to add each time (required)
  -h, --help                   help for enable
      --monthly-limit string   maximum amount in USD to auto-recharge per calendar month (required)
      --threshold string       top up when balance drops below this amount in USD (required)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic billing auto-recharge](/reference/billing/auto-recharge/auto-recharge/)	 - Manage automatic credit top-up

