---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization"
description: |-
  Get information on an Organization.
---

# Data Source: tfe_organization

Use this data source to get information about an organization.

## Example Usage

```hcl
data "tfe_organization" "foo" {
  name = "organization-name"
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - Name of the organization.
* `email` - Admin email address.
* `external_id` - An identifier for the organization.
* `assessments_enforced` - (Available only in Terraform Cloud) Whether to force health assessments (drift detection) on all eligible workspaces or allow workspaces to set thier own preferences.
* `collaborator_auth_policy` - Authentication policy (`password` or `two_factor_mandatory`). Defaults to `password`.
* `cost_estimation_enabled` - Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a Terraform Cloud organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.
* `owners_team_saml_role_id` - The name of the "owners" team.
* `send_passing_statuses_for_untriggered_speculative_plans` - Whether or not to send VCS status updates for untriggered speculative plans. This can be useful if large numbers of untriggered workspaces are exhausting request limits for connected version control service providers like GitHub. Defaults to true. In Terraform Enterprise, this setting has no effect and cannot be changed but is also available in Site Administration.
* `default_project_id` - ID of the organization's default project. All workspaces created without specifying a project ID are created in this project.