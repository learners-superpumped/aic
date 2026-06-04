---
title: "aic storage presign"
description: "Create a shareable, time-limited link"
sidebar:
  label: "presign"
---

Create a shareable, time-limited link

```
aic storage presign <bucket>/<key> [flags]
```

### Options

```
      --expires string   link lifetime, e.g. 15m, 1h (default 1h)
  -h, --help             help for presign
      --upload           create an upload link instead of download
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic storage](/reference/storage/storage/)	 - Manage AIC storage buckets and objects

