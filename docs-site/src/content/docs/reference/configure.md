---
title: "aic configure"
description: "Point aic at a non-default service (dev/self-hosted)"
sidebar:
  label: "configure"
---

Point aic at a non-default service (dev/self-hosted)

```
aic configure [flags]
```

### Options

```
      --api-endpoint string     backend API endpoint URL
      --audience-scope string   extra OIDC scope to request the API audience (provider-specific)
      --client-id string        OIDC client id for the CLI
  -h, --help                    help for configure
      --issuer string           OIDC issuer URL
      --output-format string    default output format: table|json|yaml
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic](/reference/aic/)	 - Run your company on AIC — domains, email, storage, SEO, and ads

