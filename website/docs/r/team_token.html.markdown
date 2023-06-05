---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_token"
description: |-
  Generates a new team token and overrides existing token if one exists.
---

# tfe_team_token

Generates a new team token and overrides existing token if one exists.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_token" "test" {
  team_id = tfe_team.test.id
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `force_regenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!
* `expired_at` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339 
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and 
never expire.

## Example Usage

When a token has an expiry:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_team_token" "test" {
  team_id = tfe_team.test.id
  expired_at = time_rotating.example.rotation_rfc3339
}
```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Team tokens can be imported; use `<TEAM ID>` as the import ID. For example:

```shell
terraform import tfe_team_token.test team-47qC3LmA47piVan7
```
