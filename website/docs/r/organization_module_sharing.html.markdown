---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_module_sharing"
sidebar_current: "docs-resource-tfe-organization-module-sharing"
description: |-
  Manage module sharing for an organization.
---

# tfe_organization_module_sharing

Manage module sharing for an organization.

~> **NOTE:** This resource requires using the provider with 
an instance of Terraform Enterprise at least as recent as v202004-1.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization_module_sharing" "test" {
  organization  = "my-org-name"
  module_consumers = ["my-org-name-2", "my-org-name-3"] 
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `module_consumers` - (Required) Names of the organizations to consume the module registry.
