---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_token"
description: |-
  Generates a new team token and overrides existing token if one exists.
---

# tfe_team_token

Generates a new team token. If a description is not set, then it follows the legacy behavior to override
the single team token without a description if it exists.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_token" "test" {
  team_id     = tfe_team.test.id
  description = "my team token"
}

resource "tfe_team_token" "ci" {
  team_id     = tfe_team.test.id
  description = "my second team token"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `description` - (Optional) The token's description, which must be unique per team. Required if creating multiple
  tokens for a single team.
* `expired_at` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339 
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and 
never expire.
* `force_regenerate` - (Optional) Only applies to legacy tokens without descriptions. If set to `true`, a new
  token will be generated even if a token already exists. This will invalidate the existing token! This cannot
  be set with `description`.

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
  description = "my team token"
  expired_at = time_rotating.example.rotation_rfc3339
}
```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Team tokens can be imported either by `<TOKEN ID>` or by `<TEAM ID>`. Using the team ID will follow the
legacy behavior where the imported token is the single token of the team that has no description.

For example:

```shell
terraform import tfe_team_token.test at-47qC3LmA47piVan7
terraform import tfe_team_token.test team-47qC3LmA47piVan7
```
