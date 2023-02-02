---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_policy_set"
description: |-
  Add a policy set to a workspace
---

# tfe_workspace_policy_set

Adds and removes policy sets from a workspace

-> **Note:** `tfe_policy_set` has an argument `workspace_ids` that should not be used alongside this resource. They attempt to manage the same attachments.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "Some description."
  organization  = tfe_organization.test.name
}

resource "tfe_workspace_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  workspace_id  = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `workspace_id` - (Required) Workspace ID to add the policy set to.

## Attributes Reference

* `id` - The ID of the policy set attachment. ID format: `<workspace-id>_<policy-set-id>`

## Import

Workspace Policy Sets can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>/<POLICY SET NAME>`. For example:

```shell
terraform import tfe_workspace_policy_set.test 'my-org-name/workspace/policy-set-name'
```
