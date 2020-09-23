---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_oauth_client"
sidebar_current: "docs-datasource-tfe-oauth-client-x"
description: |-
  Get information on an OAuth client.
---

# Data Source: tfe_oauth_client

Use this data source to get information about an OAuth client.

## Example Usage

```hcl
data "tfe_oauth_client" "client" {
  oauth_client_id = "oc-XXXXXXX"
}
```

## Argument Reference

The following arguments are supported:

* `oauth_client_id` - (Required) ID of the OAuth client.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The OAuth client ID. This will match `oauth_client_id`.
* `api_url` - The client's API URL.
* `http_url` - The client's HTTP URL.
* `oauth_token_id` - The ID of the OAuth token associated with the OAuth client.
* `ssh_key` - The SSH key assigned to the OAuth client.
