---
title: "aic domains records delete"
description: "Delete a DNS record set"
sidebar:
  label: "delete"
---

Delete a DNS record set

```
aic domains records delete <domain> [flags]
```

### Options

```
  -h, --help          help for delete
      --name string   record name; @ for apex
      --type string   record type (A, CNAME, MX, TXT, ...)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic domains records](/reference/domains/records/records/)	 - Manage DNS records for a connected domain

