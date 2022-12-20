---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_ssh_key"
description: |-
  Get information on a SSH key.
---

# Data Source: tfe_ssh_key

Use this data source to get information about a SSH key.

## Example Usage

```hcl
data "tfe_ssh_key" "test" {
  name         = "my-ssh-key-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the SSH key.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the SSH key.
