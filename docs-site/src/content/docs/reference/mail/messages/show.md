---
title: "aic mail messages show"
description: "Show a message's metadata; optionally write its raw .eml to a file"
sidebar:
  label: "show"
---

Show a message's metadata; optionally write its raw .eml to a file

```
aic mail messages show <message-id> [flags]
```

### Options

```
  -h, --help             help for show
      --raw-out string   write the raw .eml to this path
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

