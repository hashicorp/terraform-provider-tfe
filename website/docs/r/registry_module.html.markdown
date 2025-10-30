---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_module"
description: |-
  Manages registry modules
---

# tfe_registry_module

HCP Terraform's private module registry helps you share Terraform modules across your organization.

~> **NOTE:** The `agent_execution_mode` and `agent_pool_id` fields in the `test_config` block are currently in beta and are not available to all users. These features are subject to change or be removed.

**Note**: To manage this resource, the token used with the provider needs to be for a team with **owner** permissions or a user who has the permissions explicitly assigned. Crucially, this **does not work** with an organization token! See the [API Access Levels](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens#access-levels) documentation for more information.

## Example Usage

Basic usage with VCS:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test-oauth-client" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}

resource "tfe_registry_module" "test-registry-module" {
  vcs_repo {
    display_identifier = "my-org-name/terraform-provider-name"
    identifier         = "my-org-name/terraform-provider-name"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
  }
}
```

Create private registry module with tests enabled:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test-oauth-client" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}

resource "tfe_registry_module" "test-registry-module" {
  test_config {
    tests_enabled = true
  }

  vcs_repo {
    display_identifier = "my-org-name/terraform-provider-name"
    identifier         = "my-org-name/terraform-provider-name"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    branch             = "main"
  }
}
```

Create private registry module with agent pool (BETA):

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_oauth_client" "test-oauth-client" {
  organization     = tfe_organization.test-organization.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}

resource "tfe_registry_module" "test-registry-module" {
  test_config {
    tests_enabled         = true
    agent_execution_mode  = "agent"
    agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  }

  vcs_repo {
    display_identifier = "my-org-name/terraform-provider-name"
    identifier         = "my-org-name/terraform-provider-name"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    branch             = "main"
  }
}
```

Create private registry module with GitHub App:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

data "tfe_github_app_installation" "gha_installation" {
  name = "YOUR_GH_NAME"
}

resource "tfe_registry_module" "petstore" {
  organization = tfe_organization.test-organization.name
  vcs_repo {
    display_identifier = "GH_NAME/REPO_NAME"
    identifier         = "GH_NAME/REPO_NAME"
    github_app_installation_id     = data.tfe_github_app_installation.gha_installation.id
  }
}
```

Create private registry module without VCS:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-private-registry-module" {
  organization    = tfe_organization.test-organization.name
  module_provider = "my_provider"
  name            = "another_test_module"
  registry_name   = "private"
}
```

Create public registry module:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-public-registry-module" {
  organization    = tfe_organization.test-organization.name
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
}
```

Create no-code provisioning registry module:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-no-code-provisioning-registry-module" {
  organization    = tfe_organization.test-organization.name
  namespace       = "terraform-aws-modules"
  module_provider = "aws"
  name            = "vpc"
  registry_name   = "public"
}

resource "tfe_no_code_module" "foobar" {
  organization = tfe_organization.test-organization.id
  registry_module = tfe_registry_module.test-no-code-provisioning-registry-module.id
}
```

## Argument Reference

The following arguments are supported:

* `vcs_repo` - (Optional) Settings for the registry module's VCS repository. Forces a
  new resource if changed. One of `vcs_repo` or `module_provider` is required.
* `module_provider` - (Optional) Specifies the Terraform provider that this module is used for. For example, "aws"
* `name` - (Optional) The name of registry module. It must be set if `module_provider` is used.
* `organization` - (Optional) The name of the organization associated with the registry module. It must be set if `module_provider` is used, or if `vcs_repo` is used via a GitHub App.
* `namespace` - (Optional) The namespace of a public registry module. It can be used if `module_provider` is set and `registry_name` is public.
* `registry_name` - (Optional) Whether the registry module is private or public. It can be used if `module_provider` is set.
* `initial_version` - (Optional) This specifies the initial version for a branch based module. It can be used if `vcs_repo.branch` is set. If it is omitted, the initial modules version will default to `0.0.0`.

The `test_config` block supports:
* `tests_enabled` - (Optional) Specifies whether tests run for the registry module. Tests are only supported for branch-based publishing.
* `agent_execution_mode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) to use for registry module tests. Valid values are `agent` and `remote`. Defaults to `remote`. This feature is currently in beta and is not available to all users.
* `agent_pool_id` - (Optional) The ID of an agent pool to assign to the registry module for testing. Requires `agent_execution_mode` to be set to `agent`. This value _must not_ be provided if `agent_execution_mode` is set to `remote`. This feature is currently in beta and is not available to all users.

The `vcs_repo` block supports:

* `display_identifier` - (Required) The display identifier for your VCS repository.
  For most VCS providers outside of BitBucket Cloud and Azure DevOps, this will match the `identifier`
  string.
* `identifier` - (Required) A reference to your VCS repository in the format
  `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization (or project key, for Bitbucket Data Center)
  and repository in your VCS provider. The format for Azure DevOps is `<ado organization>/<ado project>/_git/<ado repository>`.
* `oauth_token_id` - (Optional) Token ID of the VCS Connection (OAuth Connection Token) to use. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.
* `github_app_installation_id` - (Optional) The installation id of the Github App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.
* `branch` - (Optional) The git branch used for publishing when using branch-based publishing for the registry module. When a `branch` is set, `tags` will be returned as `false`.
* `tags` - (Optional) Specifies whether tag based publishing is enabled for the registry module. When `tags` is set to `true`, the `branch` must be set to an empty value.
* `source_directory` - (Optional) The path to the module configuration files within the VCS repository. This feature is currently in beta and is not available to all users.
* `tag_prefix` - (Optional) The prefix to filter repository Git tags when using the tag-based publishing type in a repository that contains code for multiple modules. Without a prefix, HCP Terraform and Terraform Enterprise publish new versions for all modules with valid Git tags that use semantic versioning. This feature is currently in beta and is not available to all users.

## Attributes Reference

* `id` - The ID of the registry module.
* `module_provider` - The Terraform provider that this module is used for.
* `name` - The name of registry module.
* `organization` - The name of the organization associated with the registry module.
* `namespace` - The namespace of the module. For private modules this is the name of the organization that owns the module.
* `publishing_mechanism` - The publishing mechanism used when releasing new versions of the module.
* `registry_name` - The registry name of the registry module depicting whether the registry module is private or public.
* `test_config` - The test configuration for the registry module.
  * `tests_enabled` - Whether tests are enabled for the registry module.
  * `agent_execution_mode` - The execution mode used for registry module tests.
  * `agent_pool_id` - The ID of the agent pool used for registry module tests.
* `no_code` - **Deprecated** The property that will enable or disable a module as no-code provisioning ready.
Use the tfe_no_code_module resource instead.

## Import

Registry modules can be imported; use `<ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>` as the import ID. For example:

```shell
terraform import tfe_registry_module.test my-org-name/public/namespace/name/provider/mod-qV9JnKRkmtMa4zcA
```

**Deprecated** use `<ORGANIZATION NAME>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>` as the import ID. For example:

```shell
terraform import tfe_registry_module.test my-org-name/name/provider/mod-qV9JnKRkmtMa4zcA
```
