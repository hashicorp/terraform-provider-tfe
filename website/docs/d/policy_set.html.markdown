---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy_set"
sidebar_current: "docs-datasource-tfe-policy-set"
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
* `workspace_ids` - IDs of the workspaces that use the policy set.
* `policy_ids` - IDs of the policies attached to the policy set.
* `policies_path` - The sub-path within the attached VCS repository when using `vcs_repo`.
* `vcs_repo` - Settings for the workspace's VCS repository.

The `vcs_repo` block contains:

* `identifier` - A reference to your VCS repository in the format `<organization>/<repository>`
  where `<organization>` and `<repository>` refer to the organization and repository in your VCS
  provider.
* `branch` - The repository branch that Terraform will execute from.
* `ingress_submodules` - Indicates whether submodules should be fetched when
  cloning the VCS repository.
* `oauth_token_id` - OAuth token ID of the configured VCS connection.

