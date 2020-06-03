---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_notification_configuration"
sidebar_current: "docs-resource-tfe-run-trigger"
description: |-
  Manages run triggers
---

# tfe_run_trigger

Terraform Cloud provides a way to connect your workspace to one or more workspaces within your organization, known as "source workspaces". 
These connections, called run triggers, allow runs to queue automatically in your workspace on successful apply of runs in any of the source workspaces. 
You can connect your workspace to up to 20 source workspaces.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test-workspace" {
  name         = "my-workspace-name"
  organization = "${tfe_organization.test-organization.id}"
}

resource "tfe_workspace" "test-sourceable" {
  name         = "my-sourceable-workspace-name"
  organization = "${tfe_organization.test-organization.id}"
}

resource "tfe_run_trigger" "test" {
  workspace_external_id = "${tfe_workspace.test-workspace.id}"
  sourceable_id         = "${tfe_workspace.test-sourceable.id}"
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - The id of the workspace that owns the run trigger. This is the workspace where runs will be triggered.
  This value _must not_ be provided if `workspace_external_id` is provided.
* `workspace_external_id` - **Deprecated** Use `workspace_id` instead. The id of the workspace that owns the run trigger. This is the workspace where runs will be triggered.
  This value _must not_ be provided if `workspace_id` is provided.
* `sourceable_id` - (Required) The id of the sourceable. The sourceable must be a workspace.

-> **Note:** One of `workspace_id` or `workspace_external_id` _must_ be provided.

## Attributes Reference

* `id` - The ID of the run trigger.

## Import

Run triggers can be imported; use `<RUN TRIGGER ID>` as the import ID. For example:

```shell
terraform import tfe_run_trigger.test rt-qV9JnKRkmtMa4zcA
```
