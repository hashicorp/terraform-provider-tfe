---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_lock"
description: |-
  Lock Workspace.
---

# tfe_workspace_lock

[Lock](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#locking) a workspace
to prevent any runs from happening. When this resource is detroyed, the workspace is unlocked.

## Example Usage

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_workspace_lock" "test" {
  workspace_id = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `tfe_workspace` - (Required) Workspace ID
* `reason` - (Optional) Reason for locking the workspace

## Attributes Reference

* `id` - The ID of the Workspace locked.

## Import

Import an existing lock with the workspace ID.

```shell
terraform import tfe_workspace_lock.test ws-12345
```
