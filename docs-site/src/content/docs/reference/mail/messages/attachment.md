---
title: "aic mail messages attachment"
description: "Download a message attachment (id from `messages show`)"
sidebar:
  label: "attachment"
---

Download a message attachment (id from `messages show`)

```
aic mail messages attachment <message-id> <attachment-id> [flags]
```

### Options

```
  -h, --help         help for attachment
      --out string   write to this path (default: attachment filename; '-' for stdout)
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

