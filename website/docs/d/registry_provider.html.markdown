---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider"
description: |-
  Get information on a public or private provider in the private registry.
---

# Data Source: tfe_registry_provider

Use this data source to get information about a public or private provider in the private registry.

## Example Usage

A private provider:

```hcl
resource "tfe_registry_provider" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
```

A public provider:

```hcl
resource "tfe_registry_provider" "example" {
  organization  = "my-org-name"
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether this is a publicly maintained provider or private. Must be either `public` or `private`. Defaults to `private`.
* `namespace` - (Optional) The namespace of the provider. Required if `registry_name` is `public`, otherwise it can't be configured, and it will be set to same value as the `organization`.
* `name` - (Required) Name of the provider.

## Attributes Reference

* `id` - ID of the provider.
* `created_at` - The time when the provider was created.
* `updated_at` - The time when the provider was last updated.
