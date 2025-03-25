---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ssh_key"
description: |-
  Manages SSH keys.
---

# tfe_ssh_key

This resource represents an SSH key which includes a name and the SSH private
key. An organization can have multiple SSH keys available.

## Example Usage

Basic usage:

```hcl
resource "tfe_ssh_key" "test" {
  name         = "my-ssh-key-name"
  organization = "my-org-name"
  key          = "private-ssh-key"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name to identify the SSH key.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `key` - (Optional) The text of the SSH private key. One of `key` or `key_wo`
  must be provided.
* `key_wo` - (Optional) The text of the SSH private key, guaranteed not to be
  written to plan or state artifacts. One of `key` or `key_wo` must be provided.

## Attributes Reference

* `id` The ID of the SSH key.

## Import

Because the Terraform Enterprise API does not return the private SSH key
content, this resource cannot be imported.

-> **Note:** Write-Only argument `key_wo` is available to use in place of `key`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments).
