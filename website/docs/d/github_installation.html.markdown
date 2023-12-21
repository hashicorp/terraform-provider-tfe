---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_github_app_installation"
description: |-
Get information on the Github App Installation.
---

# Data Source: tfe_github_app_installation

Use this data source to get information about the Github App Installation.

## Example Usage

### Finding a Github App Installation by its installation ID

```hcl
data "tfe_github_app_installation" "gha_installation" {
  installation_id = 12345678
}
```

### Finding a Github App Installation by its name

```hcl
data "tfe_github_app_installation" "gha_installation" {
  name = "github_username_or_organization"
}
```

## Argument Reference

The following arguments are supported. At least one of `name`, `installation_id` must be set.

* `installation_id` - (Optional) ID of the Github Installation. The installation ID can be found in the URL slug when visiting the installation's configuration page, e.g `https://github.com/settings/installations/12345678`.
* `name` - (Optional) Name of the Github user or organization account that installed the app.

Must be one of: `installation_id` or `name`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The internal ID of the Github Installation. This is different from the `installation_id`.
