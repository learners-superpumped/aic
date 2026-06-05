---
title: "aic mail inboxes list"
description: "List inboxes for a domain"
sidebar:
  label: "list"
---

List inboxes for a domain

```
aic mail inboxes list [flags]
```

### Options

```
      --cursor string   next-page cursor from a previous list
      --domain string   domain to list inboxes for (required)
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

* [aic mail inboxes](/reference/mail/inboxes/inboxes/)	 - Manage sending addresses (inboxes)

