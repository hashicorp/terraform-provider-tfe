---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_settings_twilio"
sidebar_current: "docs-resource-tfe-settings-twilio"
description: |-
  Manage the Twilio settings of a Terraform Enterprise installation.
---

# tfe_settings_twilio

Manage the [Twilio settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-twilio-settings) of a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_settings_twilio" "settings" {
  enabled = true

  account_sid = "12345abcd"
  from_number = "555-555-5555"
  auth_token  = "sample_token"
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Optional) Allows Twilio to be used. If true, all remaining attributes must have valid values. Default to `false`.
* `account_sid` - (Optional) The Twilio account id.
* `from_number` - (Optional) The Twilio authentication token.
* `auth_token` - (Optional) The Twilio registered phone number that will be used to send the message.
