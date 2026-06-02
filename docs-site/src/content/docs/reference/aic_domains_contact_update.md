---
title: "aic domains contact update"
description: "Update a WHOIS contact profile (provide full set of fields)"
---

Update a WHOIS contact profile (provide full set of fields)

```
aic domains contact update <name> [flags]
```

### Options

```
      --address1 string       street address line 1
      --address2 string       street address line 2 (optional)
      --city string           city
      --country string        ISO-3166-1 alpha-2 country code (e.g. KR, US)
      --email string          registrant email
      --first-name string     registrant first name
  -h, --help                  help for update
      --last-name string      registrant last name
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

* [aic domains contact](/reference/aic_domains_contact/)	 - Manage WHOIS contact profiles for domain registration

