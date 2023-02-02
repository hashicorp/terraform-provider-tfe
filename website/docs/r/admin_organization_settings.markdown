---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_admin_organization_settings"
description: |-
  Manages admin settings for an organization (Terraform Enterprise Only).
---

# tfe_admin_organization_settings

Manage admin settings for an organization. This resource requires the
use of an admin token and is for Terraform Enterprise only. See example usage for
incorporating an admin token in your provider config.

## Example Usage

Basic usage:

```hcl

provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_organization" "a-module-producer" {
  name  = "my-org"
  email = "admin@company.com"
}

resource "tfe_organization" "a-module-consumer" {
  name  = "my-other-org"
  email = "admin@company.com"
}

resource "tfe_admin_organization_settings" "test-settings" {
  provider                              = tfe.admin
  organization                          = tfe_organization.a-module-producer.name
  workspace_limit                       = 15
  access_beta_tools                     = false
  global_module_sharing                 = false
  module_sharing_consumer_organizations = [
    tfe_organization.a-module-consumer.name
  ]
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization provider config must be defined.
* `access_beta_tools` - (Optional) True if the organization has access to beta tool versions.
* `workspace_limit` - (Optional) Maximum number of workspaces for this organization. If this number is set to a value lower than the number of workspaces the organization has, it will prevent additional workspaces from being created, but existing workspaces will not be affected. If set to 0, this limit will have no effect.
* `global_module_sharing` - (Optional) If true, modules in the organization's private module repository will be available to all other organizations. Enabling this will disable any previously configured module_sharing_consumer_organizations. Cannot be true if module_sharing_consumer_organizations is set.
* `module_sharing_consumer_organizations` - (Optional) A list of organization names to share modules in the organization's private module repository with. Cannot be set if global_module_sharing is true.

## Attributes Reference

* `sso_enabled` - True if SSO is enabled in this organization

## Import

This resource does not manage the creation of an organization and there is no need to import it.
