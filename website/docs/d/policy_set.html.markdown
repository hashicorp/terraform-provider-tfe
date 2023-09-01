---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set"
description: |-
  Get information on organization policy sets.
---

# Data Source: tfe_policy_set

This data source is used to retrieve a policy set defined in a specified organization.

## Example Usage

For workspace policies:

```hcl
data "tfe_policy_set" "test" {
  name         = "my-policy-set-name"
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy set.
* `organization` - (Required) Name of the organization.

## Attributes Reference

* `id` - The ID of the policy set.
* `organization` - Name of the organization.
* `name` - Name of the policy set.
* `description` - Description of the policy set.
* `global` - Whether or not the policy set applies to all workspaces in the organization.
* `kind` - The policy-as-code framework for the policy. Valid values are "sentinel" and "opa".
* `overridable` - Whether users can override this policy when it fails during a run. Only valid for OPA policies.
* `workspace_ids` - IDs of the workspaces that use the policy set.
* `excluded_workspace_ids` - IDs of the workspaces that do not use the policy set.
* `project_ids` - IDs of the projects that use the policy set.
* `policy_ids` - IDs of the policies attached to the policy set.
* `policies_path` - The sub-path within the attached VCS repository when using `vcs_repo`.
* `vcs_repo` - Settings for the workspace's VCS repository.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `<vcs organization>/<repository>`
  where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS
  provider.
* `branch` - The repository branch that Terraform will execute from.
* `ingress_submodules` - Indicates whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.

