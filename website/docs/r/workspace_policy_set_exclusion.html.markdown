---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_workspace_policy_set_exclusion"
description: |-
  Add a policy set to an excluded workspace
---

# tfe_workspace_policy_set_exclusion

Adds and removes policy sets from an excluded workspace

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

resource "tfe_workspace_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  workspace_id  = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `workspace_id` - (Required) Excluded workspace ID to add the policy set to.

## Attributes Reference

* `id` - The ID of the policy set attachment. ID format: `<workspace-id>_<policy-set-id>`

## Import

Excluded Workspace Policy Sets can be imported; use `<ORGANIZATION>/<WORKSPACE NAME>/<POLICY SET NAME>`. For example:

```shell
terraform import tfe_workspace_policy_set_exclusion.test 'my-org-name/workspace/policy-set-name'
```
