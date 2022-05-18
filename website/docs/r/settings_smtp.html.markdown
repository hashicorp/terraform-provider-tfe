---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_settings_smtp"
sidebar_current: "docs-resource-tfe-settings-smtp"
description: |-
  Manage the SMTP settings of a Terraform Enterprise installation.
---

# tfe_settings_smtp

Manage the [SMTP settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-smtp-settings) of a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_settings_smtp" "settings" {
  enabled = true

  host               = "example.com"
  port               = 25
  sender             = "sample_user@example.com"
  auth               = "login"
  username           = "sample_user"
  password           = "sample_password"
  test_email_address = "test@example.com"
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Optional) Allows SMTP to be used. If true, all remaining attributes must have valid values. Default to `false`.
* `host` - (Optional) The host address of the SMTP server.
* `port` - (Optional) The port of the SMTP server.
* `sender` - (Optional) The desired sender address.
* `auth` - (Optional) The authentication type. Valid values are `"none"`, `"plain"`, and `"login"`. Default to `"none"`.
* `username` - (Optional) The username used to authenticate to the SMTP server. Only required if `auth` is set to `"login"` or `"plain"`.
* `password` - (Optional) The username used to authenticate to the SMTP server. Only required if `auth` is set to `"login"` or `"plain"`.
* `test_email_address` - (Optional) The email address to send a test message to. Not persisted and only used during testing.
