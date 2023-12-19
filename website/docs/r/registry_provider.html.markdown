---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_provider"
description: |-
  Manages public and private providers in the private registry.
---

# tfe_registry_provider

Manages public and private providers in the private registry.

## Example Usage

Create private provider:

```hcl
resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  name = "my-provider"
}
```

Create public provider:

```hcl
resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

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

## Import

Providers can be imported; use `<ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>` as the import ID.

For example a private provider:

```shell
terraform import tfe_registry_provider.example my-org-name/private/my-org-name/my-provider
```

Or a public provider:

```shell
terraform import tfe_registry_provider.example my-org-name/public/hashicorp/aws
```
