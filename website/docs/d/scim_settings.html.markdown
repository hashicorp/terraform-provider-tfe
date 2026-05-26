---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_scim_settings"
description: |-
  Get information on SCIM Settings.
---

# Data Source: tfe_scim_settings

Use this data source to get information about SCIM Settings. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.

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

data "tfe_scim_settings" "foo" {
  provider = tfe.admin
}
```

## Argument Reference

No arguments are required for this data source.

## Attributes Reference

The following attributes are exported:

* `id` - It is always `scim`.
* `enabled` - Whether SCIM provisioning is enabled.
* `paused` - Whether SCIM provisioning is paused.
* `site_admin_group_scim_id` - The SCIM ID of the group mapped to site admin. Empty when no group is linked.
* `site_admin_group_display_name` - The display name of the SCIM group mapped to site admin. Empty when no group is linked.
