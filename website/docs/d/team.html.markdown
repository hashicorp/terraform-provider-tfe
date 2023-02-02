---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
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
* `sso_team_id` - (Optional) The [SSO Team ID](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) of the team, if it has been defined
