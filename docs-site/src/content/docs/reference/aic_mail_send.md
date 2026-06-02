---
title: "aic mail send"
description: "Send mail via SES"
---

Send mail via SES

```
aic mail send [flags]
```

### Options

```
      --attach strings       attachment file path (repeatable)
      --bcc strings          BCC recipient (repeatable)
      --cc strings           CC recipient (repeatable)
      --from string          sending address (must be an inbox you've created)
  -h, --help                 help for send
      --html string          HTML body
      --html-file string     read HTML body from file
      --in-reply-to string   reply to a stored message id (server sets In-Reply-To/References)
      --reply-to strings     Reply-To address (repeatable)
      --subject string       subject line
      --text string          plain-text body
      --text-file string     read plain-text body from file
      --to strings           recipient address (repeatable)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic mail](/reference/aic_mail/)	 - Manage outbound mail (SES identities, inboxes, send)

