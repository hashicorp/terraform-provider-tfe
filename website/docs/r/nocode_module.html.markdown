---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_nocode_module"
description: |-
  Manages no code for registry modules
---

# tfe_nocode_module

Creates, updates and destroys no code for registry modules.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "foobar" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
	module_provider = "my_provider"
	name            = "test_module"
}

resource "tfe_nocode_module" "foobar" {
	organization = tfe_organization.foobar.id
	module = tfe_registry_module.foobar.id
	follow_latest_version = true
	enabled = true
}
```

Creating a no-code module with variable options:

```hcl
resource "tfe_organization" "foobar" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "foobar" {
	organization    = tfe_organization.foobar.id
	module_provider = "my_provider"
	name            = "test_module"
}

resource "tfe_nocode_module" "foobar" {
	organization = tfe_organization.foobar.id
	module = tfe_registry_module.foobar.id
	follow_latest_version = true
	enabled = true

	variable_options {
		name    = "ami"
		type    = "string"
		options = [ "ami-0", "ami-1", "ami-2" ]
	}

	variable_options {
		name    = "region"
		type    = "string"
		options = [ "us-east-1", "us-east-2", "us-west-1"]
	}
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the variable set.
- `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
- `module` - (Required) The ID of the registry module to associate with the no code module.
- `enabled` - (Required) Whether or not no-code module is enabled for the associated registry module
- `follow_latest_version` - (Optional) Whether or not to follow the latest version of the module. Must be `false` if `version_pin` is set.
- `version_pin` - (Optional) The version of the module to pin to. Must not be set if `follow_latest_version` is `true`.
- `variable_options` - (Optional) A list of variable options to associate with the no code module.
  - `name` - (Required) The name of the variable option.
  - `type` - (Required) The type of the variable option.
  - `options` - (Required) A list of options for the variable option.

## Attributes Reference

- `id` - The ID of the no code module.

## Import

No-code modules can be imported; use `<NO CODE MODULE ID>` as the import ID. For example:

```shell
terraform import tfe_nocode_module.test nocode-qV9JnKRkmtMa4zcA
```
