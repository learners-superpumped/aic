---
title: "aic seo search-console query"
description: "Search analytics"
sidebar:
  label: "query"
---

Search analytics

```
aic seo search-console query <domain> [flags]
```

### Options

```
      --dimensions strings   query|page|country|device
      --end string           end date YYYY-MM-DD
      --filter stringArray   dimension filter "<dim> <op> <expr>", e.g. "country equals usa" (repeatable)
  -h, --help                 help for query
      --limit int            row limit
      --start string         start date YYYY-MM-DD
      --start-row int        pagination offset (0 = first page)
      --type string          web|image|video|news|discover|googleNews (default web)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic seo search-console](/reference/seo/search-console/search-console/)	 - Query Search Console data

