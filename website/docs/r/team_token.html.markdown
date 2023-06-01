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
* `expired_at` - (Optional) A date and time in which the team token will expire. The expiration date must be passed in
iso8601 format. If no expiration date is supplied, the expiration date will default to null and never expire.

## Example Usage

Basic usage:
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
  expired_at = time_rotating.example.id
}
```

Generating the `expired_at` string using the date tool in unix systems (darwin):
```
date -Iseconds -v"+30d"
```

Generating the `expired_at` string using the date tool in unix systems (linux):
```
date -Iseconds -d"+30 days"
```

Generating the `expired_at` string using the `timeadd` Terraform function:
```
$ terraform console
> timeadd(timestamp(), "720h")
"2023-07-21T02:02:23Z"
```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Team tokens can be imported; use `<TEAM ID>` as the import ID. For example:

```shell
terraform import tfe_team_token.test team-47qC3LmA47piVan7
```
