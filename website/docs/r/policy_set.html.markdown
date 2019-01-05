---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set"
sidebar_current: "docs-resource-tfe-tfe_policy_set"
description: |-
  Sentinel Policy as Code is an embedded policy as code framework integrated with Terraform Enterprise.
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
  name                   = "my-policy-set"
  description            = "A brand new policy set"
  organization           = "my-org-name"
  policy_ids             = ["${tfe_sentinel_policy.test.id}"]
  workspace_external_ids = ["${tfe_workspace.test.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy set.
* `description` - (Optional) A description of the policy set's purpose.
* `global` - (Optional) Whether or not policies in this set will apply to
  all workspaces. Defaults to `false`. This value _must not_ be provided if
  `workspace_external_ids` are provided.
* `organization` - (Required) Name of the organization.
* `policy_ids` - (Required) A list of Sentinel policy IDs.
* `workspace_external_ids` - (Optional) A list of workspace external IDs. If
  the policy set is `global`, this value _must not_ be provided.

## Attributes Reference

* `id` - The ID of the policy set.

## Import

Policy sets can be imported; use `<POLICY SET ID>` as the import ID. For example:

```shell
terraform import tfe_policy_set.test polset-wAs3zYmWAhYK7peR
```
