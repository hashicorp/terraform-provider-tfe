---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_gpg_keys"
description: |-
  Get information on private registry GPG keys of an organization.
---

# Data Source: tfe_registry_gpg_key

Use this data source to get information about all private registry GPG keys of an organization.

## Example Usage

```hcl
data "tfe_registry_gpg_keys" "all" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

* `keys` - List of GPG keys in the organization. Each element contains the following attributes:
  * `id` - ID of the GPG key.
  * `organization` - Name of the organization.
  * `ascii_armor` - ASCII-armored representation of the GPG key.
  * `created_at` - The time when the GPG key was created.
  * `updated_at` - The time when the GPG key was last updated.
