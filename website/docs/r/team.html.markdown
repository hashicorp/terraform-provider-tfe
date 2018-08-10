---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
sidebar_current: "docs-tfe-team"
description: |-
  Creates and destroys teams.
---

# tfe_team

Creates and destroys teams.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "team" {
  name = "my-team-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the team.
* `organization` - (Required) Name of the organization.

## Attributes Reference

* `id` The ID of the team.
