---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_registry_module"
sidebar_current: "docs-resource-tfe-registry-module"
description: |-
  Manages registry modules
---

# tfe_registry_module

Terraform Cloud's private module registry helps you share Terraform modules across your organization. 

## Example Usage

Basic usage:

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

## Argument Reference

The following arguments are supported:

* `vcs_repo` - (Required) Settings for the registry module's VCS repository. Forces a
  new resource if changed.

The `vcs_repo` block supports:

* `display_identifier` - (Required) The display identifier for your VCS repository.
   For most VCS providers outside of BitBucket Cloud, this will match the `identifier` 
   string.
* `identifier` - (Required) A reference to your VCS repository in the format
  `<organization>/<repository>` where `<organization>` and `<repository>` refer to the organization (or project key, for Bitbucket Server) 
  and repository in your VCS provider. The format for Azure DevOps is <organization>/<project>/_git/<repository>.
* `oauth_token_id` - (Required) Token ID of the VCS Connection (OAuth Connection Token)
  to use.

## Attributes Reference

* `id` - The ID of the registry module.
* `module_provider` - The provider of the registry module.
* `name` - The name of registry module.
* `organization` - The name of the organization associated with the registry module.

## Import

Registry modules can be imported; use `<ORGANIZATION NAME>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>` as the import ID. For example:

```shell
terraform import tfe_registry_module.test my-org-name/name/provider/mod-qV9JnKRkmtMa4zcA
```
