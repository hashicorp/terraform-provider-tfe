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
* `sso_team_id` - (Optional) The [SSO Team ID](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) of the team, if it has been defined.
* `scim_linked` - Whether the team is linked to a SCIM group. Only populated when SCIM is enabled on the TFE instance.
* `scim_group_name` - The display name of the SCIM group linked to this team. Only populated when SCIM is enabled and the team is linked to a SCIM group.
* `scim_sync_paused` - Whether SCIM membership sync is paused for this team. Only populated when SCIM is enabled and the team is linked to a SCIM group.
* `scim_updated_at` - The timestamp of the last SCIM reconciliation for this team, in RFC3339 format. Only populated when SCIM is enabled and the team is linked to a SCIM group.
