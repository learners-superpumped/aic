---
title: "aic ads update"
description: "Edit a running ad campaign (budget, targeting, end date)"
sidebar:
  label: "update"
---

Edit a running ad campaign (budget, targeting, end date)

```
aic ads update <campaign-id> [flags]
```

### Options

```
      --age string               replace target age range, e.g. 25-44
      --budget int               new budget in nano-dollars
      --end string               new campaign end time (RFC3339)
      --genders stringArray      replace target genders: male|female (repeatable)
      --geo stringArray          replace target country/region codes (repeatable)
  -h, --help                     help for update
      --interests stringArray    replace Meta interest IDs (repeatable)
      --placements stringArray   replace placements (repeatable)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic ads](/reference/ads/ads/)	 - Manage AIC ad campaigns

