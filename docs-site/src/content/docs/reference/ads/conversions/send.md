---
title: "aic ads conversions send"
description: "Relay a Meta CAPI events payload through AIC"
sidebar:
  label: "send"
---

Relay a Meta CAPI events payload through AIC

### Synopsis

Reads a Meta Conversions API JSON body ({"data":[...]}) from --file or stdin
and relays it to Meta for this project's pixel. PII (em, ph, ...) must be
SHA-256 hashed by you per Meta's rules — AIC forwards the payload as-is.

```
aic ads conversions send [flags]
```

### Options

```
      --file string   path to a Meta CAPI JSON body (default: stdin)
  -h, --help          help for send
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic ads conversions](/reference/ads/conversions/conversions/)	 - Send server-side conversion events (Meta CAPI passthrough)

