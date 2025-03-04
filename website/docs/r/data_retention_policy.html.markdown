---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_data_retention_policy"
description: |-
  Manages data retention policies for organizations and workspaces
---

# tfe_data_retention_policy

Creates a data retention policy attached to either an organization or workspace. This resource is for Terraform Enterprise only.

## Example Usage

Creating a data retention policy for a workspace:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.test-workspace.id

  delete_older_than {
    days = 42
  }
}
```

Creating a data retention policy for an organization:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}
```

Creating a data retention policy for an organization and exclude a single workspace from it:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// create data retention policy the organization
resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = tfe_organization.test-organization.name
}

// create a policy that prevents automatic deletion of data in the test-workspace
resource "tfe_data_retention_policy" "foobar" {
  workspace_id = tfe_workspace.test-workspace.id

  dont_delete {}
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) The name of the organization you want the policy to apply to. Must not be set if `workspace_id` is set.
* `workspace_id` - (Optional) The ID of the workspace you want the policy to apply to. Must not be set if `organization` is set.
* `delete_older_than` - (Optional) If this block is set, the created policy will apply to any data older than the configured number of days. Must not be set if `dont_delete` is set.
* `dont_delete` - (Optional) If this block is set, the created policy will prevent other policies from deleting data from this workspace or organization. Must not be set if `delete_older_than` is set.


## Import

A resource can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>` or `<ORGANIZATION>` as the import ID. For example:

```shell
terraform import tfe_data_retention_policy.foobar my-org-name/my-workspace-name
```
