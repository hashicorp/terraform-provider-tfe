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

Basic usage (VCS-based policy set):

```hcl
resource "tfe_policy_set" "test" {
  name                   = "my-policy-set"
  description            = "A brand new policy set"
  organization           = "my-org-name"
  workspace_external_ids = ["${tfe_workspace.test.external_id}"]

  vcs_repo {
    identifier         = "my-org-name/my-policy-set-repository"
    branch             = "master"
    ingress_submodules = false
    oauth_token_id     = "${tfe_oauth_client.test.oauth_token_id}"
  }

  policies_path = "policies/my-policy-set"
}
```

Using manually-specified policies:

```hcl
resource "tfe_policy_set" "test" {
  name                   = "my-policy-set"
  description            = "A brand new policy set"
  organization           = "my-org-name"
  policy_ids             = ["${tfe_sentinel_policy.test.id}"]
  workspace_external_ids = ["${tfe_workspace.test.external_id}"]
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
* `workspace_external_ids` - (Optional) A list of workspace external IDs. If
  the policy set is `global`, this value _must not_ be provided.
* `vcs_repo` - (Optional) The [VCS repository
  settings](#vcs-repository-settings) for this policy set. Forces a new resource
  if changed.
* `policies_path` - (Optional) The sub-path within the attached VCS repository
  to ingress when using `vcs_repo`. All files and directories outside of this
  sub-path will be ignored. This option can only be supplied when `vcs_repo` is
  present. Forces a new resource if changed.
* `policy_ids` - (Optional) A list of Sentinel policy IDs.

-> **Note:** When neither `vcs_repo` or `policy_ids` is not specified, the current default
is to create an empty non-VCS policy set.

### VCS Repository Settings

The `vcs_repo` block takes the following arguments:

* `identifier` - (Required) The identifier of the VCS repository in the format
  `<namespace>/<repo>`. For example, on GitHub, this would be something like
  `hashicorp/my-policy-set`.
* `oauth_token_id` - (Required) The ID of the OAuth token assocaited with the
  VCS repository to use. This can be fetched from the `oauth_token_id` attribute
  of a `tfe_oauth_client` resource.
* `branch` - (Optional) The branch of the VCS repo. If empty, the VCS provider's
  default branch value will be used.
* `ingress_submodules` - (Optional) Determines whether repository submodules
  will be instantiated during the clone operation. Default: `false`.
 
## Attributes Reference

* `id` - The ID of the policy set.

## Import

Policy sets can be imported; use `<POLICY SET ID>` as the import ID. For example:

```shell
terraform import tfe_policy_set.test polset-wAs3zYmWAhYK7peR
```
