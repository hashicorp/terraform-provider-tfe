---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_no_code_module"
description: |-
  Get information on a no-code module.
---

# Data Source: tfe_registry_provider

Use this data source to read the details of an existing No-Code-Allowed module.

## Example Usage

```hcl
resource "tfe_no_code_module" "foobar" {
	organization = tfe_organization.foobar.id
	registry_module = tfe_registry_module.foobar.id
}

data "tfe_no_code_module" "foobar" {
	id = tfe_no_code_module.foobar.id
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) ID of the no-code module. 

## Attributes Reference

* `id` - ID of the no-code module.
* `organization` - Organization name that the no-code module belongs to.
* `namespace` - Namespace name that the no-code module belongs to.
* `registry_module_id` - ID of the registry module for the no-code module. 
* `version_pin` - Version number the no-code module is pinned to.
* `enabled` - Indicates if this no-code module is currently enabled
