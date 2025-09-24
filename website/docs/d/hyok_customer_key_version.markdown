---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_hyok_customer_key_version"
description: |-
  Get information on a HYOK customer key version.
---

# Data Source: tfe_hyok_customer_key_version

Use this data source to get information about a Hold Your Own Keys (HYOK) customer key version.

## Example Usage

```hcl
data "tfe_hyok_customer_key_version" "tfe_hyok_customer_key_version1" {
  id = "keyv-123"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the HYOK customer key version.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_at` - The time when the customer key version was created.
* `error` - Any error message associated with the customer key version.
* `id` - The ID of the customer key version.
* `key_version` - The version number of the customer key.
* `status` - The status of the customer key version.
* `workspaces_secured` - The number of workspaces securefd by this customer key version.
