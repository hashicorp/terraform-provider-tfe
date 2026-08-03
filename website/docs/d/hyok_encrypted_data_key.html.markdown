---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_hyok_encrypted_data_key"
description: |-
  Get information on a HYOK encrypted data key.
---

# Data Source: tfe_hyok_encrypted_data_key

Use this data source to get information about a Hold Your Own Keys (HYOK) encrypted data key.

## Example Usage

```hcl
data "tfe_hyok_encrypted_data_key" "tfe_hyok_encrypted_data_key1" {
  id = "dek-123"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the HYOK encrypted data key.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_at` - The time when the encrypted data key was created.
* `customer_key_name` - The name of the customer key used to encrypt the data key.
* `encrypted_dek` - The encrypted data encryption key (DEK).
* `id` - The ID of the encrypted data key.