---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_client"
description: |-
  Manages OAuth clients.
---

# tfe_oauth_client

An OAuth Client represents the connection between an organization and a VCS
provider.

-> **Note:** This resource does not currently support creation of Azure DevOps Services OAuth clients.

## Example Usage

Basic usage:

```hcl
resource "tfe_oauth_client" "test" {
  name             = "my-github-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-vcs-provider-token"
  service_provider = "github"
}
```

#### Azure DevOps Server Usage

See [documentation for TFC/E setup](https://developer.hashicorp.com/terraform/cloud-docs/vcs/azure-devops-server).

**Note:** This resource requires a private key when creating Azure DevOps Server OAuth clients.

```hcl
resource "tfe_oauth_client" "test" {
  name             = "my-ado-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://ado.example.com"
  http_url         = "https://ado.example.com"
  oauth_token      = "my-vcs-provider-token"
  private_key      = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  service_provider = "ado_server"
}
```

#### BitBucket Server Usage

See [documentation for TFC/E setup](https://developer.hashicorp.com/terraform/cloud-docs/vcs/bitbucket-server).

When using BitBucket Server, you must use three required fields: `key`, `secret`, `rsa_public_key`.


```hcl
resource "tfe_oauth_client" "test" {
  name             = "my-bbs-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://bbs.example.com"
  http_url         = "https://bss.example.com"
  key              = "<consumer key>"
  secret           = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  rsa_public_key   = "-----BEGIN PUBLIC KEY-----\ncontent\n-----END PUBLIC KEY-----"
  service_provider = "bitbucket_server"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Display name for the OAuth Client. Defaults to the `service_provider` if not supplied.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `api_url` - (Required) The base URL of your VCS provider's API (e.g.
  `https://api.github.com` or `https://ghe.example.com/api/v3`).
* `http_url` - (Required) The homepage of your VCS provider (e.g.
  `https://github.com` or `https://ghe.example.com`).
* `oauth_token` - The token string you were given by your VCS provider, e.g. `ghp_xxxxxxxxxxxxxxx` for a GitHub personal access token. For more information on how to generate this token string for your VCS provider, see the [Create an OAuth Client](https://developer.hashicorp.com/terraform/cloud-docs/api-docs/oauth-clients#create-an-oauth-client) documentation.
* `private_key` - (Required for `ado_server`) The text of the private key associated with your Azure DevOps Server account
* `key` - The OAuth Client key can refer to a Consumer Key, Application Key,
  or another type of client key for the VCS provider.
* `secret` - (Required for `bitbucket_server`) The OAuth Client secret is used for BitBucket Server, this secret is the
  the text of the SSH private key associated with your BitBucket Server
Application Link.
* `rsa_public_key` - (Required for `bitbucket_server`) Required for BitBucket
  Server in conjunction with the secret. Not used for any other providers. The
text of the SSH public key associated with your BitBucket Server Application
Link.
* `service_provider` - (Required) The VCS provider being connected with. Valid
  options are `ado_server`, `ado_services`, `bitbucket_hosted`, `bitbucket_server`, `github`, `github_enterprise`, `gitlab_hosted`,
  `gitlab_community_edition`, or `gitlab_enterprise_edition`.

## Attributes Reference

* `id` - The ID of the OAuth client.
* `oauth_token_id` - The ID of the OAuth token associated with the OAuth client.
