---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_clients"
sidebar_current: "docs-datasource-tfe-oauth-clients"
description: |-
  Get Terraform Cloud/Enterprise's oauth clients in a given  organization.
---

# Data Source: tfe_oauth_clients

Use this data source to retrieve a list of TerraformCloud/Enterprise's oauth clients in a given organization.

## Example Usage

```hcl
data "tfe_oauth_clients"  "example" {
  organization = "my-org-name"
}

output "oauth_clients" {
  value       = data.tfe_oauth_clients.example.oauth_clients
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `oauth_clients` - A list of oauth clients manage by the organization. Structure is documented below.

### oauth_clients
* `oauth_client_id` - The oauth client ID.
* `api_url` - The base URL of your VCS provider's API.
* `callback_url` - The callback URL.
* `connect_path` - The connect path.
* `created_at` - The creation date.
* `http_url` - The homepage of your VCS provider.
* `key` - The oauth client key.
* `service_provider` - The VCS provider being connected with.
* `service_provider_display_name` - The display name of your VCS provider.
* `organization_name` -  Name of the organization.
* `oauth_token_id` - The ID of the oauth token associated with the oauth client.
