---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_module"
description: |-
  Manages registry modules
---

# tfe_registry_module

Terraform Cloud's private module registry helps you share Terraform modules across your organization.

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

The `vcs_repo` block supports:

* `display_identifier` - (Required) The display identifier for your VCS repository.
  For most VCS providers outside of BitBucket Cloud, this will match the `identifier`
  string.
* `identifier` - (Required) A reference to your VCS repository in the format
  `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization (or project key, for Bitbucket Server)
  and repository in your VCS provider. The format for Azure DevOps is <organization>/<project>/\_git/<repository>.
* `oauth_token_id` - (Optional) Token ID of the VCS Connection (OAuth Connection Token) to use. This conflicts with `github_app_installation_id` and can only be used if `github_app_installation_id` is not used.
* `github_app_installation_id` - (Optional) The installation id of the Github App. This conflicts with `oauth_token_id` and can only be used if `oauth_token_id` is not used.

## Attributes Reference

* `id` - The ID of the registry module.
* `module_provider` - The Terraform provider that this module is used for.
* `name` - The name of registry module.
* `organization` - The name of the organization associated with the registry module.
* `namespace` - The namespace of the module. For private modules this is the name of the organization that owns the module.
* `registry_name` - The registry name of the registry module depicting whether the registry module is private or public.
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
