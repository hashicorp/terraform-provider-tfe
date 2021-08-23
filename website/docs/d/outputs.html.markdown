---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_outputs"
sidebar_current: "docs-datasource-tfe-state-outputs"
description: |-
  Get output values from another organization/workspace.
---
# Data Source: tfe_outputs

This data source is used to retrieve the state outputs for a given workspace.
It enables output values in one Terraform configuration to be used in another.

The outputs retrieved from this data source may contain sensitive information.
To that end, we defaulted to setting the `values` attribute — which contains the
output data — of this data source to be marked as
[sensitive()](https://www.terraform.io/docs/language/functions/sensitive.html).
This means that one must use `nonsensitive()` to display the output value.

## Example Usage

Using the `tfe_outputs` data source, the outputs `foo` and `bar` can be used as seen below:

In the example below, assume we have outputs defined in an my-org/my-workspace:

```
output "foo" {
  value = "a"
}

output "bar" {
  value = "b"
}
```

The `tfe_outputs` data source can now use `foo` and `bar`
dynamically as seen below.

```hcl
data "tfe_outputs" "foobar" {
  organization = "my-org"
  workspace = "my-workspae"
}

output "hello" {
	value = data.tfe_outputs.foobar.values.foo
}

output "world" {
	value = data.tfe_outputs.foobar.values.bar
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) The name of the organization.
* `workspace` - (Required) The name of the workspace.

## Attributes Reference

The following attributes are exported:

* `values` - The current output values for the specified workspace.
