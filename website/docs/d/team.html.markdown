---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
sidebar_current: "docs-datasource-tfe-team-x"
description: |-
  Get information on a team.
---

# Data Source: tfe_team

Use this data source to get information about a team.

## Example Usage

```hcl
data "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the team.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the team.
