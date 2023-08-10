---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_teams"
description: |-
  Get information on Teams.
---

# Data Source: tfe_teams

Use this data source to get a list of Teams in an Organization and a map of their IDs. The Teams returned may be a subset of all teams in an Organization based on the permissions of the API token.

## Example Usage

```hcl
data "tfe_teams" "foo" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:
* `id` - Name of the organization.
* `names` - A list of team names in an organization.
* `ids` - A map of team names in an organization and their IDs.