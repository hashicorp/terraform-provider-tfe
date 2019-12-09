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

-> **Note:** This resource does not currently support creation of Bitbucket
  Server OAuth clients.

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

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.
* `api_url` - (Required) The base URL of your VCS provider's API (e.g.
  `https://api.github.com` or `https://ghe.example.com/api/v3`).
* `http_url` - (Required) The homepage of your VCS provider (e.g.
  `https://github.com` or `https://ghe.example.com`).
* `oauth_token` - (Required) The token string you were given by your VCS provider.
* `service_provider` - (Required) The VCS provider being connected with. Valid
  options are `ado_server`, `ado_services`, `github`, `github_enterprise`, `gitlab_hosted`,
  `gitlab_community_edition`, or `gitlab_enterprise_edition`.
* `private_key` - (Optional) The text of the private key associated with your VCS provider user account

-> **Note:** `private_key` is only available when the `service_provder` is set to Azure DevOps Server (`ado_server`)


## Attributes Reference

* `id` - The ID of the OAuth client.
* `oauth_token_id` - The ID of the OAuth token associated with the OAuth client.
