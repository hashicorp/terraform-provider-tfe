---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_state_outputs"
sidebar_current: "docs-datasource-tfe-state-outputs"
description: |-
  Retrieves the State outputs per organization and workspace.
---
# Data Source: tfe_state_outputs

This data source is used to retrieve the state outputs for a given workspace.
It enables values in the outputs to be used dynamically in a terraform
configuration.

## Example Usage

Using the `tfe_state_outputs` data source in a terraform configuration.

In the example below, assume we have a state outputs that looks like this:

```
{
  "version": <version>,
  "terraform_version": "<terraform-version>",
  ...
  "outputs": {
    "identifier": {
      "value": "9023256633839603543",
      "type": "string"
    },
    "records": {
      "value": ["hashicorp.com", "terraform.io"],
      "type": ["list", "string"]
    }
  },
  "resources": [
    ...
  ]
}
```

The `tfe_state_outputs` data source can now use `identifier` and `records`
dynamically as seen below.

```hcl
data "tfe_state_outputs" "foobar" {
  organization = "<organization-name>"
  workspace = "<workspace-name>"
}

output "identifier" {
	value = data.tfe_state_outputs.foobar.values.identifier
}

output "records" {
	value = data.tfe_state_outputs.foobar.values.records
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) The name of the organizatin.
* `workspace` - (Required) The name of the workspace.

## Attributes Reference

The following attributes are exported:

* `values` - A dynamic value that can call any state outputs key and retrieve
  its value.
