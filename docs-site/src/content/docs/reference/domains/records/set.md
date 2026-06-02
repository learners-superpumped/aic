---
title: "aic domains records set"
description: "Create or replace a DNS record set (UPSERT)"
sidebar:
  label: "set"
---

Create or replace a DNS record set (UPSERT)

```
aic domains records set <domain> [flags]
```

### Options

```
  -h, --help                help for set
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

* [aic domains records](/reference/domains/records/records/)	 - Manage DNS records for a connected domain

