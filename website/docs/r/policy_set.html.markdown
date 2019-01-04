---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set"
sidebar_current: "docs-resource-tfe-tfe_policy_set"
description: |-
  Manages policy sets.
---

# tfe_policy_set

Sentinel Policy as Code is an embedded policy as code framework integrated
with Terraform Enterprise.

Policy sets are groups of policies that are applied together to related workspaces.
By using policy sets, you can group your policies by attributes such as environment
or region. Individual policies that are members of policy sets will only be checked
for workspaces that the policy set is attached to.

## Example Usage

Basic usage:

```hcl
resource "tfe_policy_set" "test" {
  name          = "my-policy-set-name"
  organization  = "my-org-name"
  policy_ids    = ["pol-X1eNTk7E9s8etcF2"]
  workspace_ids = ["my-org-name/my-workspace-name"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy set.
* `description` - (Optional) A description of the policy set's purpose.
* `organization` - (Required) Name of the organization.
* `global` - (Optional) Whether or not policies in this set will apply to
  all workspaces. Defaults to `false`.
* `policy_ids` - (Required) A list of Sentinel policy IDs to add to the set.
* `workspace_ids` - (Optional) A list of workspace IDs which should be attached
  to the policy set.
* `workspace_external_ids` - (**Deprecated**) This attribute is deprecated,
  please use the `workspace_ids` attribute instead.

## Attributes Reference

* `id` - The ID of the policy set.

## Import

Policy sets can be imported; use `<POLICY SET ID>` as the import ID. For example:

```shell
terraform import tfe_policy_set.test polset-wAs3zYmWAhYK7peR
```
