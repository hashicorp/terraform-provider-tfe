---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_provider_set"
description: |-
  Retrieve provider set.
---

# Data Source: tfe_provider_set

Retrieve a provider set by name.

~> **NOTE:** This data source is currently in beta and isn't generally
available to all users. It is subject to change or be removed.


## Example Usage

Basic usage:

```hcl
data "tfe_provider_set" "my_provider_set" {
  name         = "example-provider-set"
  organization = "example-org"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the provider set.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

* `id` - The ID of the provider set.
* `provider_source` -  Source address of the provider, e.g. `registry.terraform.io/hashicorp/tfe`.
* `description` -  Description of the provider set.
* `global` -  Whether the provider set applies globally.
* `workspace_ids` -  IDs of the workspaces attached to the provider set.
* `project_ids` -  IDs of the projects attached to the provider set.

