---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_client"
sidebar_current: "docs-resource-tfe-oauth-client"
description: |-
  Manages OAuth clients.
---

# tfe_oauth_client

An OAuth Client represents the connection between an organization and a VCS
provider.

-> **Note:** This resource does not currently support creation of Bitbucket Cloud, 
  Bitbucket Server, or Azure DevOps Services OAuth clients.

## Example Usage

Basic usage:

```hcl
resource "tfe_oauth_client" "test" {
  organization     = "my-org-name"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}
```

Azure DevOps Server usage:
-> **Note:** This resource requires a private key when creating Azure DevOps Server OAuth clients.

```hcl
resource "tfe_oauth_client" "test" {
  organization     = "my-org-name"
  api_url          = "https://ado.example.com"
  http_url         = "https://ado.example.com"
  oauth_token      = "my-vcs-provider-token"
  private_key      = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  service_provider = "ado_server"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `api_url` - (Required) The base URL of your VCS provider's API (e.g.
  `https://api.github.com` or `https://ghe.example.com/api/v3`).
* `http_url` - (Required) The homepage of your VCS provider (e.g.
  `https://github.com` or `https://ghe.example.com`).
* `oauth_token` - (Required) The token string you were given by your VCS provider.
* `private_key` - (Required for `ado_server`) The text of the private key associated with your OAuth client. Required for cloning Git submodules.
* `service_provider` - (Required) The VCS provider being connected with. Valid
  options are `ado_server`, `ado_services`, `github`, `github_enterprise`, `gitlab_hosted`,
  `gitlab_community_edition`, or `gitlab_enterprise_edition`.

## Attributes Reference

* `id` - The ID of the OAuth client.
* `oauth_token_id` - The ID of the OAuth token associated with the OAuth client.
