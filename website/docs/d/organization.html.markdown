---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization"
sidebar_current: "docs-datasource-tfe-organization"
description: |-
  Get information on an Organization.
---

# Data Source: tfe_organization

Use this data source to get information about an organization.

## Example Usage

```hcl
data "tfe_organizations" "foo" {
  name = "organization-name"
}
```

## Argument Reference

The following arguments are supported:
* `name` - (Required) Name of the organization.

## Attributes Reference

The following attributes are exported:

* `name` - Name of the organization.
* `email` - Admin email address.
* `external_id` - An identifier for the organization.
* `collaborator_auth_policy` - Authentication policy (`password`
  or `two_factor_mandatory`). Defaults to `password`.
* `cost_estimation_enabled` - Whether or not the cost estimation feature is enabled for all workspaces in the organization. Defaults to true. In a Terraform Cloud organization which does not have Teams & Governance features, this value is always false and cannot be changed. In Terraform Enterprise, Cost Estimation must also be enabled in Site Administration.
* `enterprise_plan` - Represents the Enterprise Plan type.
* `owners_team_saml_role_id` - The name of the "owners" team.
* `permissions` - Represents the organization permissions
* `session_timeout_minutes` - Session timeout after inactivity. Defaults to `20160`.
* `session_remember_minutes` - Session expiration. Defaults to `20160`.
