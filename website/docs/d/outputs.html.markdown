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

As the outputs retrieved from a different workspace may contain sensitive
information in the context of that workspace, the `values` attribute of this
data source is statically marked as
[sensitive](https://www.terraform.io/docs/language/values/outputs.html#sensitive-suppressing-values-in-cli-output).

## Example Usage

Using the `tfe_outputs` data source, the outputs `foo` and `bar` can be used as seen below:

In the example below, assume we have outputs defined in an my-org/my-workspace:

```hcl
data "tfe_outputs" "foo" {
  organization = "my-org"
  workspace = "my-workspace"
}

resource "random_id" "vpc_id" {
  keepers = {
    # Generate a new ID any time the value of 'bar' in workspace 'my-org/my-workspace' changes.
    bar = data.tfe_outputs.foo.values.bar
  }

  byte_length = 8
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) The name of the organization.
* `workspace` - (Required) The name of the workspace.

## Attributes Reference

The following attributes are exported:

* `values` - The current output values for the specified workspace.
