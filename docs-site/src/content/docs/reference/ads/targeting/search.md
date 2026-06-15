---
title: "aic ads targeting search"
description: "Search the targeting catalog (interests, behaviors, …) for ids to target"
sidebar:
  label: "search"
---

Search the targeting catalog (interests, behaviors, …) for ids to target

```
aic ads targeting search [flags]
```

### Options

```
      --dimension string   catalog: interests|behaviors|life_events|demographics|locales|geo (default "interests")
  -h, --help               help for search
      --limit int          max results (1-50) (default 25)
  -q, --query string       free-text search query (required)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic ads targeting](/reference/ads/targeting/targeting/)	 - Look up targeting options

