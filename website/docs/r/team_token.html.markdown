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
resource "tfe_team" "test" {
  name = "my-team-name"
  organization = "my-org-name"
}

resource "tfe_team_token" "test" {
  team_id = "${tfe_team.test.id}"
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

## Import

Team tokens can be imported with an ID of `<TEAM ID>`. For example:

```shell
terraform import tfe_team_token.test team-47qC3LmA47piVan7
```
