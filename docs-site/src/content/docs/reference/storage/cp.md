---
title: "aic storage cp"
description: "Upload or download an object"
sidebar:
  label: "cp"
---

Upload or download an object

### Synopsis

Upload:   aic storage cp ./file.txt <bucket>/<key>
Download: aic storage cp <bucket>/<key> ./file.txt

```
aic storage cp <src> <dst> [flags]
```

### Options

```
  -h, --help   help for cp
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

