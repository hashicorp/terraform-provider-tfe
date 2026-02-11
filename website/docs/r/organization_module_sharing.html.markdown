---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_module_sharing"
description: |-
  Manage module sharing for an organization.
---

# tfe_organization_module_sharing

Manage module sharing for an organization. This resource requires the
use of an admin token and is for Terraform Enterprise only.

-> **NOTE:** `tfe_organization_module_sharing` is deprecated in favor of using `tfe_admin_organization_settings` which also allows the management of the global module sharing setting. They attempt to manage the same resource and are mutually exclusive.

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

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `module_consumers` - (Required) Names of the organizations to consume the module registry.
