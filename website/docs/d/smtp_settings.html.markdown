---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_smtp_settings"
description: |-
  Get information on SMTP Settings.
---

# Data Source: tfe_smtp_settings

Use this data source to get information about SMTP Settings. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.


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

data "tfe_smtp_settings" "foo" {
  provider = tfe.admin
}
```

## Argument Reference

No arguments are required for this data source.

## Attributes Reference

The following attributes are exported:

* `id` - It is always `smtp`.
* `enabled` - Whether SMTP is enabled.
* `host` - The hostname of the SMTP server.
* `port` - The port of the SMTP server.
* `sender` - The sender email address.
* `auth` - The authentication type. Valid values are `none`, `plain`, and `login`.
* `username` - The username used to authenticate to the SMTP server.