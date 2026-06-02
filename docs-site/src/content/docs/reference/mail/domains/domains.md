---
title: "aic mail domains"
description: "Manage SES domain identities"
sidebar:
  label: "Overview"
  order: 0
---

Manage SES domain identities

### Options

```
  -h, --help   help for domains
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic mail](/reference/mail/mail/)	 - Manage outbound mail (SES identities, inboxes, send)
* [aic mail domains disable](/reference/mail/domains/disable/)	 - Disable a domain (deletes identity + DKIM records)
* [aic mail domains enable](/reference/mail/domains/enable/)	 - Enable a domain for outbound mail (SES identity + DKIM)
* [aic mail domains list](/reference/mail/domains/list/)	 - List email-enabled domains
* [aic mail domains show](/reference/mail/domains/show/)	 - Show identity status and DNS records
* [aic mail domains verify](/reference/mail/domains/verify/)	 - Force an immediate verification re-check

