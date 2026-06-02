---
title: "aic domains records add"
description: "Add a DNS record (fails if it already exists)"
---

Add a DNS record (fails if it already exists)

```
aic domains records add <domain> [flags]
```

### Options

```
  -h, --help                help for add
      --name string         record name; @ for apex
      --ttl int32           TTL in seconds (default 300)
      --type string         record type (A, AAAA, CNAME, MX, TXT, CAA, SRV)
      --value stringArray   record value (repeatable for multi-value sets)
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic domains records](/reference/aic_domains_records/)	 - Manage DNS records for a connected domain

