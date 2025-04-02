---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_module"
description: |-
  Get information on a registry module
---

# Data Source: tfe_registry_module

Use this data source to get information about a registry module.

## Example Usage

Basic usage:

Since modules have a [required naming convention](https://developer.hashicorp.com/terraform/registry/modules/publish#requirements), you can get these values from your module repository (`terraform-<module_provider>-<name>`). 

```hcl
data "tfe_registry_module" "example" {
  organization    = var.organization_name
  name            = "no-code-ssm"
  module_provider = "aws"
}
```

## Argument Reference

The following arguments are required:

* `organization` - (Required) The name of the organization associated with the registry module.
* `name` - (Required) The name of the module. Can be found from repository name convention `terraform-<provider>-<name>.`
* `module_provider` - (Required) The provider associated with the module. Can be found from repository name convention `terraform-<provider>-<name>.`

The following arguments are supported:

* `namespace` - (Optional) The namespace of a registry module. Typically this is the same as the organization name. Defaults to `organization` value.
* `registry_name` - (Optional) The registry name of a registry module. Valid options: "public", "private". Defaults to "private".

## Attributes Reference

* `id` - The ID of the registry module.
* `no_code_module_id` - The ID of the no-code module, if enabled.
* `no_code_module_source` - The source value of the no-code module (`<ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>`).
* `no_code` - Boolean value if no-code module is enabled.
* `publishing_mechanism` - The publishing mechanism used when releasing new versions of the module.
* `vcs_repo` - Settings for the registry module's VCS repository. 
* `permissions` - The permissions on a module.
* `status` - Current status of registry module.
* `test_config` - Test configuration indicating module testing setup.
* `version_statuses` - Version information for a given module.
* `created_at` - Date module was created.
* `updated_at` - Date module was last updated.

The `vcs_repo` block supports:

* `display_identifier` - The display identifier for your VCS repository.
  For most VCS providers outside of BitBucket Cloud and Azure DevOps, this will match the `identifier`
  string.
* `identifier` - A reference to your VCS repository in the format
  `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization (or project key, for Bitbucket Data Center)
  and repository in your VCS provider. The format for Azure DevOps is `<ado organization>/<ado project>/_git/<ado repository>`.
* `oauth_token_id` - Token ID of the VCS Connection (OAuth Connection Token) to use. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.
* `github_app_installation_id` - The installation id of the Github App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.
* `branch` - The git branch used for publishing when using branch-based publishing for the registry module. When a `branch` is set, `tags` will be returned as `false`.
* `tags` - Specifies whether tag based publishing is enabled for the registry module. When `tags` is set to `true`, the `branch` must be set to an empty value.

The `permissions` block supports:

- `can_delete` - Can delete.
- `can_resync` - Can resync.
- `can_retry` -  Can retry.

The `test_config` block supports:

- `tests_enabled` - Indicates whether tests are enabled for a module

The `version_statuses` block supports:

- `version` - Version of the module.
- `status` - Status of the module at specific version.
- `error` - Error message reported by module at specific version.
