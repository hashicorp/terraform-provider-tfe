---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization"
description: |-
  Manages organizations.
---

# tfe_organization

Manages organizations.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the organization.
* `email` - (Required) Admin email address.
* `session_timeout_minutes` - (Optional) Session timeout after inactivity.
  Defaults to `20160`.
* `session_remember_minutes` - (Optional) Session expiration. Defaults to
  `20160`.
* `collaborator_auth_policy` - (Optional) Authentication policy (`password`
  or `two_factor_mandatory`). Defaults to `password`.
* `owners_team_saml_role_id` - (Optional) The name of the "owners" team.
* `cost_estimation_enabled` - (Optional) Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a Terraform Cloud organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.
* `send_passing_statuses_for_untriggered_speculative_plans` - (Optional) Whether or not to send VCS status updates for untriggered speculative plans. This can be useful if large numbers of untriggered workspaces are exhausting request limits for connected version control service providers like GitHub. Defaults to false. In Terraform Enterprise, this setting has no effect and cannot be changed but is also available in Site Administration.
* `assessments_enforced` - (Optional) (Available only in Terraform Cloud) Whether to force health assessments (drift detection) on all eligible workspaces or allow workspaces to set their own preferences.
* `allow_force_delete_workspaces` - (Optional) Whether workspace administrators are permitted to delete workspaces with resources under management. If false, only organization owners may delete these workspaces. Defaults to false.

## Attributes Reference

* `id` - The name of the organization.

## Import

Organizations can be imported; use `<ORGANIZATION NAME>` as the import ID. For
example:

```shell
terraform import tfe_organization.test my-org-name
```
