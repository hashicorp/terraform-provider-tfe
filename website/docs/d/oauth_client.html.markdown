---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_client"
description: |-
  Get information on an OAuth client.
---

# Data Source: tfe_oauth_client

Use this data source to get information about an OAuth client.

## Example Usage

### Finding an OAuth client by its ID

```hcl
data "tfe_oauth_client" "client" {
  oauth_client_id = "oc-XXXXXXX"
}
```

### Finding an OAuth client by its name

```hcl
data "tfe_oauth_client" "client" {
  organization = "my-org"
  name         = "my-oauth-client"
}
```

### Finding an OAuth client by its service provider

```hcl
data "tfe_oauth_client" "client" {
  organization     = "my-org"
  service_provider = "github"
}
```

## Argument Reference

The following arguments are supported. At least one of `name`, `oauth_client_id`,
or `service_provider` must be set. `name` and `service_provider` may be used
together. If either `name` or `service_provider` is set, `organization` must also
be set.

* `name` - (Optional) Name of the OAuth client.
* `oauth_client_id` - (Optional) ID of the OAuth client.
* `organization` - (Optional) The name of the organization in which to search.
* `service_provider` - (Optional) The API identifier of the OAuth service provider. If set,
  must be one of: `ado_server`, `ado_services`, `bitbucket_hosted`, `bitbucket_server`,
  `github`, `github_enterprise`, `gitlab_hosted`, `gitlab_community_edition`, or
  `gitlab_enterprise_edition`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The OAuth client ID. This will match `oauth_client_id`.
* `api_url` - The client's API URL.
* `callback_url` - OAuth callback URL to provide to the OAuth service provider.
* `created_at` - The date and time this OAuth client was created in RFC3339 format.
* `http_url` - The client's HTTP URL.
* `oauth_token_id` - The ID of the OAuth token associated with the OAuth client.
* `name` - The name of the OAuth client (may be `null`).
* `organization` - The organization in which the OAuth client is registered.
* `service_provider` - The API identifier of the OAuth service provider.
* `service_provider_display_name` - The display name of the OAuth service provider.
