---
title: "aic ads launch"
description: "Launch an ad campaign"
sidebar:
  label: "launch"
---

Launch an ad campaign

```
aic ads launch [flags]
```

### Options

```
      --age string                target age range, e.g. 25-44
      --body string               ad body copy
      --budget int                budget in nano-dollars (1 USD = 1 000 000 000)
      --budget-type string        budget type: daily|lifetime (default "daily")
      --creative-asset string     storage reference for the creative asset (bucket/key)
      --cta string                call-to-action label, e.g. 'Learn More'
      --custom-audience string    Meta custom audience ID to target
      --end string                campaign end time (RFC3339); required for lifetime budgets
      --genders stringArray       target genders: male|female (repeatable)
      --geo stringArray           target country/region codes, e.g. KR US (repeatable)
      --headline string           ad headline
  -h, --help                      help for launch
      --interests stringArray     Meta interest IDs, e.g. 6003107902433 (repeatable)
      --launch-token string       idempotency token (auto-generated if omitted)
      --no-wait                   return immediately with the draft instead of waiting for the ad to be prepared
      --objective string          campaign objective: traffic|conversions|awareness|engagement|leads (default "traffic")
      --placements stringArray    ad placements, e.g. facebook_feed instagram_stories (repeatable)
      --provider string           ad provider (only 'meta' is supported) (default "meta")
      --provider-options string   provider-specific options as a JSON object
      --start string              campaign start time (RFC3339); defaults to now
      --url string                destination URL for the ad
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

