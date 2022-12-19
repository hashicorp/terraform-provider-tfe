---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_admin_settings_general"
description: |-
  Manage the general settings of a Terraform Enterprise installation.
---

# tfe_admin_settings_general

Manage the [general settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-general-settings) of a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_admin_settings_general" "settings" {
  limit_user_organization_creation                        = true
  api_rate_limiting_enabled                               = true
  api_rate_limit                                          = 30
  send_passing_statuses_for_untriggered_speculative_plans = false
  allow_speculative_plans_on_pull_requests_from_forks     = false
  default_remote_state_access                             = true
}
```

## Argument Reference

The following arguments are supported:

* `limit_user_organization_creation` - (Optional) When set to `true`, limits the ability to create organizations to users with the `site-admin` permission only. Default to `true`.
* `api_rate_limiting_enabled` - (Optional) Whether or not rate limiting is enabled for API requests. Default to `true`.
* `api_rate_limit` - (Optional) The number of allowable API requests per second for any client. Default to 30.
* `send_passing_statuses_for_untriggered_speculative_plans` - (Optional) When set to `true`, workspaces automatically send passing commit statuses for any pull requests that don't affect their tracked files. Default to `false`.
* `allow_speculative_plans_on_pull_requests_from_forks` - (Optional) When set to `false`, speculative plans are not run on pull requests from forks of a repository. Default to `false`.
* `default_remote_state_access` - (Optional) Determines the default value for the `global-remote-state` attribute on new workspaces. Default to `true`.
