---
title: "aic domains contact create"
description: "Create a WHOIS contact profile"
sidebar:
  label: "create"
---

Create a WHOIS contact profile

```
aic domains contact create --name=<name> [flags]
```

### Options

```
      --address1 string       street address line 1
      --address2 string       street address line 2 (optional)
      --city string           city
      --country string        ISO-3166-1 alpha-2 country code (e.g. KR, US)
      --default               mark this profile as the team's default (first profile auto-becomes default)
      --email string          registrant email
      --first-name string     registrant first name
  -h, --help                  help for create
      --last-name string      registrant last name
      --name string           profile name (unique per team; e.g. default, client-acme) [required]
      --organization string   company name (presence => ContactType=Company)
      --phone string          E.164 dot-notation (e.g. +82.1012345678)
      --state string          state/region (required for US/CA)
      --zip string            postal code
```

### Options inherited from parent commands

```
  -o, --output string    output format: table|json|yaml (default "table")
      --profile string   credentials profile to use (default "default")
  -p, --project string   target project (overrides the default project)
  -t, --team string      target team (overrides the default team)
```

### SEE ALSO

* [aic domains contact](/reference/domains/contact/contact/)	 - Manage WHOIS contact profiles for domain registration

