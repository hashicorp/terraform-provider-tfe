---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_tags"
description: |-
  Get information on an organization's workspace tags.
---

# Data Source: tfe_organization_tags

Use this data source to get information about the workspace tags for a given organization.

## Example Usage

```hcl
data "tfe_organization_tags" "example" {
  organization = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `tags` - A list of workspace tags within the organization

The `tag` block contains:

* `name` - The name of the workspace tag
* `id` - The ID of the workspace tag
* `workspace_count` - The number of workspaces the tag is associate with