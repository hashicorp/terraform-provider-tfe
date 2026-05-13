---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_smtp_settings"
description: |-
  Manages SMTP Settings.
---

# tfe_smtp_settings

Use this resource to create, update and destroy SMTP Settings. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.

## Example Usage

Basic usage for SMTP Settings without authentication:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_smtp_settings" "this" {
  host   = "smtp.example.com"
  port   = 25
  sender = "noreply@example.com"
  auth   = "none"
}
```

With authentication using plain password:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_smtp_settings" "this" {
  host     = "smtp.example.com"
  port     = 587
  sender   = "noreply@example.com"
  auth     = "plain"
  username = "smtp_user"
  password = "smtp_password"
}
```

With write-only password:

```hcl
variable "smtp_password" {
  type      = string
  ephemeral = true
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_smtp_settings" "this" {
  host                = "smtp.example.com"
  port                = 587
  sender              = "noreply@example.com"
  auth                = "login"
  username            = "smtp_user"
  password_wo         = var.smtp_password
  password_wo_version = 1
}
```

## Argument Reference

The following arguments are supported:

* `host` - (Optional) The hostname of the SMTP server.
* `port` - (Optional) The port of the SMTP server. Defaults to `25`.
* `sender` - (Optional) The desired sender email address.
* `auth` - (Optional) The authentication type. Valid values are `none`, `plain`, and `login`. Defaults to `none`.
* `username` - (Optional) The username used to authenticate to the SMTP server. Required if auth is `login` or `plain`.
* `password` - (Optional) The password used to authenticate to the SMTP server. Required if auth is `login` or `plain`. Cannot be used with `password_wo`.
* `password_wo` - (Optional, [Write-Only](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments)) The password used to authenticate to the SMTP server, guaranteed not to be written to plan or state artifacts. Either `password` or `password_wo` can be provided, but not both. Must be used with `password_wo_version`.
* `password_wo_version` - (Optional) Version of the write-only password. This field is used to trigger updates when the write-only password changes. Must be used with `password_wo`. When `password_wo_version` changes, the write-only password will be updated.
* `test_email_address` - (Optional) The email address to send a test message to. This value is not persisted and is only used during testing.

-> **Note:** Write-Only argument `password_wo` is available to use in place of `password`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments).

## Attributes Reference

* `id` - The ID of the SMTP settings. Always `smtp`.
* `enabled` - Whether SMTP is enabled. When enabled, all other attributes must have valid values.

## Import

SMTP Settings can be imported.

```shell
terraform import tfe_smtp_settings.this smtp
