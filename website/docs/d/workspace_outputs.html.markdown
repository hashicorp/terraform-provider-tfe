---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_outputs"
sidebar_current: "docs-datasource-tfe-state-outputs"
description: |-
  Retrieves the State outputs per organization and workspace.
---
# Data Source: tfe_workspace_outputs

This data source is used to retrieve the state outputs for a given workspace.
It enables values in the outputs to be used dynamically in a terraform
configuration.

## Example Usage

Using the `tfe_workspace_outputs` data source in a terraform configuration.

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
    },
    "secret": {
      "value": "token",
      "type": "string",
      "sensitive": true
    }
  },
  "resources": [
    ...
  ]
}
```

The `tfe_workspace_outputs` data source can now use `identifier` and `records`
dynamically as seen below.

```hcl
data "tfe_workspace_outputs" "foobar" {
  organization = "<organization-name>"
  workspace = "<workspace-name>"
}

output "identifier" {
	value = data.tfe_workspace_outputs.foobar.values.identifier
}

output "records" {
	value = data.tfe_workspace_outputs.foobar.values.records
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

* `organization` - (Required) The name of the organizatin.
* `workspace` - (Required) The name of the workspace.
* `sensitive` - (Optional) Determines whether or not to show sensitive values.
  Set to `true` to reveal sensitive values.

## Attributes Reference

The following attributes are exported:

* `values` - A dynamic value that can call any state outputs key and retrieve
  its value.
