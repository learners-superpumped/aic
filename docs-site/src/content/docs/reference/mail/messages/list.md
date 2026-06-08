---
title: "aic mail messages list"
description: "List stored messages (newest first)"
sidebar:
  label: "list"
---

List stored messages (newest first)

```
aic mail messages list [flags]
```

### Options

```
      --cursor string      next page's cursor — the 'next_cursor' value from a previous list (shown in -o json output and the table footer)
      --direction string   filter: sent|received
      --from string        filter by sender address
  -h, --help               help for list
      --inbox string       filter by inbox address or id
      --limit int          max rows per page (default 50, max 200)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic mail messages](/reference/mail/messages/messages/)	 - List and read stored mail messages

