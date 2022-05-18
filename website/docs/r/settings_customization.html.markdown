---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_settings_customization"
sidebar_current: "docs-resource-tfe-settings-customization"
description: |-
  Manage customization settings for a Terraform Enterprise installation.
---

# tfe_settings_customization

Manage [customization settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-customization-settings) for a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_settings_customization" "settings" {
  support_email_address = "support@hashicorp.com"
  login_help            = ""
  footer                = ""
  error                 = ""
  new_user              = ""
}
```

## Argument Reference

The following arguments are supported:

* `support_email_address` - (Optional) The support address for outgoing emails. Default to `"support@hashicorp.com"`.
* `login_help` - (Optional) The login help text presented to users on the login page. Default to `""`.
* `footer` - (Optional) Custom footer content that is added to the application. Default to `""`.
* `error` - (Optional) Error instruction content that is presented to users upon unexpected errors. Default to `""`.
* `new_user` - (Optional) New user instructions that is presented when the user is not yet attached to an organization. Default to `""`.
