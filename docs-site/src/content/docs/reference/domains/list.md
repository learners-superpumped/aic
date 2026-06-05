---
title: "aic domains list"
description: "List domains in the current project"
sidebar:
  label: "list"
---

List domains in the current project

```
aic domains list [flags]
```

### Options

```
      --cursor string   next page's cursor — the 'next_cursor' value from a previous list (shown in -o json output and the table footer)
  -h, --help            help for list
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

* [aic domains](/reference/domains/domains/)	 - Search, buy, and renew domains

