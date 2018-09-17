---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
sidebar_current: "docs-resource-tfe-team-x"
description: |-
  Manages teams.
---

# tfe_team

Manages teams.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
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

## Import

Teams can be imported by the `resource id`, e.g.

```shell
terraform import tfe_team.test team-47qC3LmA47piVan7
```
