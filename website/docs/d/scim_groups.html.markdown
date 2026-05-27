---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_scim_groups"
description: |-
  Get information on SCIM groups synchronized from an Identity Provider.
---

# Data Source: tfe_scim_groups

Use this data source to read SCIM groups that have been synchronized from the
configured Identity Provider into Terraform Enterprise. It applies only to
Terraform Enterprise and requires admin token configuration. See example usage
for incorporating an admin token in your provider config.

Exactly one of `name` or `search` must be provided:

* Use `name` to look up a single group by its exact display name
  (case-insensitive). The data source filters out fuzzy substring matches
  returned by the API and keeps only exact matches.
* Use `search` to retrieve every group whose name matches the API's substring
  query (`?q=<search>`). Results across all pages are returned.

## Example Usage

Look up a single SCIM group by its exact name and reference its ID:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_scim_groups" "admins" {
  provider = tfe.admin
  name     = "platform-admins"
}

output "admin_group_id" {
  value = data.tfe_scim_groups.admins.group_id
}
```

List every SCIM group whose name contains a given substring:

```hcl
data "tfe_scim_groups" "engineering" {
  provider = tfe.admin
  search   = "-eng-"
}

output "engineering_group_names" {
  value = [for g in data.tfe_scim_groups.engineering.groups : g.name]
}
```

Pair with `tfe_scim_settings` to map a SCIM group to the site admin role:

```hcl
data "tfe_scim_groups" "site_admins" {
  provider = tfe.admin
  name     = "tfe-site-admins"
}

resource "tfe_scim_settings" "this" {
  provider                 = tfe.admin
  site_admin_group_scim_id = data.tfe_scim_groups.site_admins.group_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The exact name of the SCIM group to retrieve
  (case-insensitive). Cannot be used with `search`.
* `search` - (Optional) A substring used to filter SCIM groups by name via the
  API's query parameter (`?q=<search>`, case-insensitive). Cannot be used with `name`.

## Attributes Reference

The following attributes are exported:

* `id` - The internal ID of the data source, formatted as `<argument>/<value>`
  (e.g., `name/platform-admins` or `search/-eng-`). The `<value>` portion is
  URL-path-escaped, so characters such as spaces or `/` appear percent-encoded
  (e.g., `name/platform%20admins`).
* `groups` - The list of all matching SCIM groups. Each entry exports:
    * `id` - The ID of the SCIM group.
    * `name` - The name of the SCIM group.
* `group_id` - The ID of the SCIM group. Only populated when exactly one
  matching group is found; otherwise null.
* `group_name` - The name of the SCIM group. Only populated when exactly one
  matching group is found; otherwise null.
