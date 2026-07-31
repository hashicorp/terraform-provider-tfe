---
layout: "tfe"
page_title: "Terraform Enterprise: Resource tfe_registry_module"
description: |-
  Manages Terraform modules in an organization's private module registry.
  ~> Warning: The agent_execution_mode and agent_pool_id fields in the test_config block are currently in beta and are not available to all users.
  ~> Note:  To manage this resource, the token used with the provider needs to be for a team with owner permissions or a user who has the permissions explicitly assigned. Crucially, this does not work with an organization token! See the API Access Levels https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens#access-levels documentation for more information.
  ~> Note: When using source_directory, you must explicitly specify both name and module_provider. This is required because monorepos and repositories with non-standard names (not following terraform-<provider>-<name> convention) cannot have these values automatically inferred by the API.
---

# Resource: tfe_registry_module

Manages Terraform modules in an organization's private module registry.

~> **Warning:** The `agent_execution_mode` and `agent_pool_id` fields in the `test_config` block are currently in beta and are not available to all users.

~> **Note:**  To manage this resource, the token used with the provider needs to be for a team with **owner** permissions or a user who has the permissions explicitly assigned. Crucially, this **does not work** with an organization token! See the [API Access Levels](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens#access-levels) documentation for more information.

~> **Note:** When using `source_directory`, you **must** explicitly specify both `name` and `module_provider`. This is required because monorepos and repositories with non-standard names (not following `terraform-<provider>-<name>` convention) cannot have these values automatically inferred by the API.

## Example Usage

```terraform
# Basic usage with VCS

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

```terraform
# Create private registry module with tests enabled

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

```terraform
# Create private registry module with agent pool (BETA)

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
    tests_enabled        = true
    agent_execution_mode = "agent"
    agent_pool_id        = tfe_agent_pool.test-agent-pool.id
  }

  vcs_repo {
    display_identifier = "my-org-name/terraform-provider-name"
    identifier         = "my-org-name/terraform-provider-name"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    branch             = "main"
  }
}
```

```terraform
# Create private registry module with GitHub App

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
    display_identifier         = "GH_NAME/REPO_NAME"
    identifier                 = "GH_NAME/REPO_NAME"
    github_app_installation_id = data.tfe_github_app_installation.gha_installation.id
  }
}
```

```terraform
# Create private registry module from a monorepo with source_directory (BETA)

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

resource "tfe_registry_module" "monorepo-module" {
  organization    = tfe_organization.test-organization.name
  name            = "vpc"
  module_provider = "aws"

  vcs_repo {
    display_identifier = "my-org-name/private-modules"
    identifier         = "my-org-name/private-modules"
    oauth_token_id     = tfe_oauth_client.test-oauth-client.oauth_token_id
    source_directory   = "modules/vpc"
  }
}
```

```terraform
# Create private registry module without VCS

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

```terraform
# Create public registry module

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

```terraform
# Create no-code provisioning registry module

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
  organization    = tfe_organization.test-organization.id
  registry_module = tfe_registry_module.test-no-code-provisioning-registry-module.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `initial_version` (String) The initial version for a branch-based module. Can be used if `vcs_repo.branch` is set. Defaults to `0.0.0`.
- `module_provider` (String) Specifies the Terraform provider that this module is used for. For example, `aws`. Required with `name`.
- `name` (String) The name of the registry module. Must be set if `module_provider` is used.
- `namespace` (String) The namespace of a public registry module. Can be used if `module_provider` is set and `registry_name` is `public`.
- `no_code` (Boolean, Deprecated) Whether to enable no-code provisioning for this module. **Deprecation notes**: Use the `tfe_no_code_module` resource instead.
- `organization` (String) The name of the organization associated with the registry module. Must be set if `module_provider` is used, or if `vcs_repo` is used via a GitHub App. If omitted, organization must be defined in the provider config.
- `registry_name` (String) Whether the registry module is `private` or `public`. Can be used if `module_provider` is set.
- `test_config` (Block List) Settings for running tests for the registry module. (see [below for nested schema](#nestedblock--test_config))
- `vcs_repo` (Block List, Max: 1) Settings for the registry module's VCS repository. One of `vcs_repo` or `module_provider` is required. (see [below for nested schema](#nestedblock--vcs_repo))

### Read-Only

- `id` (String) The ID of the registry module.
- `publishing_mechanism` (String) The publishing mechanism used when releasing new versions of the module.

<a id="nestedblock--test_config"></a>
### Nested Schema for `test_config`

Optional:

- `agent_execution_mode` (String) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode) to use for registry module tests. Valid values are `agent` and `remote`. Defaults to `remote`. This feature is currently in beta and is not available to all users.
- `agent_pool_id` (String) The ID of an agent pool to use for registry module tests. Requires `agent_execution_mode` to be `agent`. Beta feature, not available to all users.
- `tests_enabled` (Boolean) Whether tests are enabled for the registry module. Tests are only supported for branch-based publishing.


<a id="nestedblock--vcs_repo"></a>
### Nested Schema for `vcs_repo`

Required:

- `identifier` (String) A reference to your VCS repository in the format `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization (or project key, for Bitbucket Data Center) and repository in your VCS provider. The format for Azure DevOps is `<ado organization>/<ado project>/_git/<ado repository>`. Changes to this field update the module in place.

Optional:

- `branch` (String) The git branch used for publishing when using branch-based publishing. When set, `tags` will be returned as `false`. Changes to this field update the module in place.
- `display_identifier` (String) The display identifier for your VCS repository. For most VCS providers outside of BitBucket Cloud and Azure DevOps, this will match the `identifier` string. HCP Terraform recomputes this server-side for OAuth connections; it is read back from the API and does not need to be set explicitly.
- `github_app_installation_id` (String) The installation ID of the GitHub App. Conflicts with `oauth_token_id`. Changes to this field update the module in place. Switching from `github_app_installation_id` to `oauth_token_id` is supported.
- `oauth_token_id` (String) Token ID of the VCS Connection (OAuth Connection Token) to use. Conflicts with `github_app_installation_id`. Changes to this field update the module in place. Switching from `oauth_token_id` to `github_app_installation_id` is supported.
- `source_directory` (String) The path to the module configuration files within the VCS repository. Changes to this field update the module in place. Beta feature, not available to all users.
- `tag_prefix` (String) The prefix to filter repository Git tags when using tag-based publishing in a repository that contains code for multiple modules. Without a prefix, HCP Terraform and Terraform Enterprise publish new versions for all modules with valid Git tags that use semantic versioning. Beta feature, not available to all users.
- `tags` (Boolean) Whether tag-based publishing is enabled for the registry module. When `true`, `branch` must be set to an empty value.




## Import

tfe_registry_module can be imported using an identity. For example:

```terraform
import {
  to = tfe_registry_module.test
  identity = {
    id              = "mod-qV9JnKRkmtMa4zcA"
    organization    = "my-org-name"
    registry_name   = "private"
    namespace       = "my-org-name"
    name            = "vpc"
    module_provider = "aws"
    hostname        = "app.terraform.io"
  }
}
```

<!-- schema generated by tfplugindocs -->
### Identity Schema

#### Required

- `id` (String)
- `module_provider` (String)
- `name` (String)
- `namespace` (String)
- `organization` (String)
- `registry_name` (String)

#### Optional

- `hostname` (String)


Resource tfe_registry_module can be imported in the following format: 

```shell
# via <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>
terraform import tfe_registry_module.test my-org-name/public/namespace/name/provider/mod-qV9JnKRkmtMa4zcA

# **Deprecated** via <ORGANIZATION NAME>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>
terraform import tfe_registry_module.test my-org-name/name/provider/mod-qV9JnKRkmtMa4zcA
```
