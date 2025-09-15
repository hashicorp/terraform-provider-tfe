---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_stack"
description: |-
  Manages Stacks.
---

# tfe_stack

Defines a Stack resource.

~> **NOTE:** Stacks support in the hashicorp/tfe provider is currently available on a pre-release basis and should be considered beta software and subject to change. One notable aspect of this resource is that a stack may not be destroyed until all resources within its deployments have been destroyed.

## Example Usage

### Create a stack with a VCS repository:

```hcl
resource "tfe_oauth_client" "test" {
  organization     = "my-example-org"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = var.github_token
  service_provider = "github"
}

data "tfe_organization" "organization" {
  name = "my-example-org"
}

data "tfe_agent_pool" "agent-pool" {
  name                  = "my-example-agent-pool"
  organization          = tfe_organization.organization.name
}

resource "tfe_stack" "test-stack" {
  name          = "my-stack"
  description   = "A Terraform Stack using two components with two environments"
  project_id    = data.tfe_organization.organization.default_project_id
  agent_pool_id = data.tfe_agent_pool.agent-pool.id

  vcs_repo {
    branch         = "main"
    identifier     = "my-github-org/stack-repo"
    oauth_token_id = tfe_oauth_client.test.oauth_token_id
  }
}
```

### Create a stack without a VCS repository:

```hcl
resource "tfe_oauth_client" "test" {
  organization     = "my-example-org"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = var.github_token
  service_provider = "github"
}

data "tfe_organization" "organization" {
  name = "my-example-org"
}

resource "tfe_stack" "test-stack" {
  name         = "my-stack"
  description  = "A Terraform Stack using two components with two environments"
  project_id   = data.tfe_organization.organization.default_project_id
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the stack.
* `project_id` - (Required) ID of the project where the stack should be created.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the stack.
* `vcs_repo` - (Optional) Settings for the stack's VCS repository.
* `description` - (Optional) Description of the stack
<!--
NOTE: This is a proposed schema for allowing force-delete actions on a stack. Force delete is not implemented yet so I've commented it out for now.

* `force_delete` - (Optional) If this argument is true, the stack will be deleted during destroy plans even if it contains deployments that have managed resources. You may need to apply this change to the stack before running terraform destroy. Without this argument, all resources managed by stacks deployments need to be destroyed before the stack may be destroyed.-->

The `vcs_repo` block supports:

* `identifier` - (Required) A reference to your VCS repository in the format `<vcs organization>/<repository>` where `<vcs organization>` and `<repository>` refer to the organization and repository in your VCS provider. The format for Azure DevOps is `<ado organization>/<ado project>/_git/<ado repository>`.
* `branch` - (Optional) The repository branch that Terraform will execute from. This defaults to the repository's default branch (e.g. main).
* `github_app_installation_id` - (Optional) The installation id of the Github App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.
* `oauth_token_id` - (Optional) The VCS Connection (OAuth Connection + Token) to use. This ID can be obtained from a `tfe_oauth_client` resource. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.

## Attributes Reference

* `id` - The stack ID.
* `deployment_names` - The set of deployment names used in the last configuration for this stack. This attribute will be empty when the resource is created and will remain empty until a configuration is fetched.

## Import

Stacks can be imported by ID, which can be found on the stack's settings tab in the UI

Example:

```shell
terraform import tfe_stack.test-stack st-9cs9Vf6Z49Zzrk1t
```
