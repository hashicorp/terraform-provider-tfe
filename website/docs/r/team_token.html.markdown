---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_token"
sidebar_current: "docs-resource-tfe-team-token"
description: |-
  Generates a new team token and overrides existing token if one exists.
---

# tfe_team_token

Generates a new team token and overrides existing token if one exists.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "team" {
  name = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_token" "token" {
  team_id = "${tfe_team.team.id}"
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team.
* `force_regenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.
