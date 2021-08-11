---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_outputs"
sidebar_current: "docs-datasource-tfe-state-outputs"
description: |-
  Get output values from another organization/workspace.
---
# Data Source: tfe_workspace_outputs

This data source is used to retrieve the state outputs for a given workspace.
It enables output values in one Terraform configuration to be used in another.

## Example Usage

Using the `tfe_workspace_outputs` data source, the outputs `foo` and `bar` can be used as seen below:

In the example below, assume we have outputs defined in an my-org/my-workspace:

```
output "foo" {
  value = "a"
}

output "bar" {
  value = "b"
}
```

The `tfe_workspace_outputs` data source can now use `foo` and `bar`
dynamically as seen below.

```hcl
data "tfe_workspace_outputs" "foobar" {
  organization = "my-org"
  workspace = "my-workspae"
}

output "hello" {
	value = data.tfe_workspace_outputs.foobar.values.foo
}

output "world" {
	value = data.tfe_workspace_outputs.foobar.values.bar
}
```

If you want to reveal sensitive values, then set the optional boolean flag
`sensitive=true`:

```
data "tfe_workspace_outputs" "foobar" {
  organization = "<organization-name>"
  workspace = "<workspace-name>"
  sensitive = true
}

output "secret" {
	value = data.tfe_workspace_outputs.foobar.values.secret
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) The name of the organization.
* `workspace` - (Required) The name of the workspace.
* `sensitive` - (Optional) Determines whether or not to show sensitive values.
  Set to `true` to reveal sensitive values.

## Attributes Reference

The following attributes are exported:

* `values` - The current output values for the specified workspace.
