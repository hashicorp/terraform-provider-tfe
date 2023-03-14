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
  installation_id = 12345
}
```

### Finding a Github App Installation by its name

```hcl
data "tfe_github_app_installation" "gha_installation" {
  name = "installation_name"
}
```

## Argument Reference

The following arguments are supported. At least one of `name`, `installation_id` must be set. 

* `installation_id` - (Optional) ID of the Github Installation as shown in Github.
* `name` - (Optional) Name of the Github Installation as shown in Github.
 
Must be one of: `installation_id` or `name`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The internal ID of the Github Installation. This is different from the `installation_id`.