---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_providers"
description: |-
  Get information on public and private providers in the private registry.
---

# Data Source: tfe_registry_providers

Use this data source to get information about public and private providers in the private registry.

## Example Usage

All providers:

```hcl
data "tfe_registry_providers" "all" {
  organization = "my-org-name"
}
```

All private providers:

```hcl
data "tfe_registry_providers" "private" {
  organization  = "my-org-name"
  registry_name = "private"
}
```

Providers with "hashicorp" in their namespace or name:

```hcl
data "tfe_registry_providers" "hashicorp" {
  organization  = "my-org-name"
  search        = "hashicorp"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `registry_name` - (Optional) Whether to list only public or private providers. Must be either `public` or `private`.
* `search` - (Optional) A query string to do a fuzzy search on provider name and namespace.

## Attributes Reference

* `providers` - List of the providers. Each element contains the following attributes:
  * `id` - ID of the provider.
  * `organization` - Name of the organization.
  * `registry_name` - Whether this is a publicly maintained provider or private.
  * `namespace` - Namespace of the provider.
  * `name` -  Name of the provider.
  * `created_at` - Time when the provider was created.
  * `updated_at` - Time when the provider was last updated.
