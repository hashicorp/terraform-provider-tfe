---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_gpg_key"
description: |-
  Get information on a private registry GPG key.
---

# Data Source: tfe_registry_gpg_key

Use this data source to get information about a private registry GPG key.

## Example Usage

```hcl
data "tfe_registry_gpg_key" "example" {
  organization = "my-org-name"
  id           = "13DFECCA3B58CE4A"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) ID of the GPG key.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

* `ascii_armor` - ASCII-armored representation of the GPG key.
* `created_at` - The time when the GPG key was created.
* `updated_at` - The time when the GPG key was last updated.
