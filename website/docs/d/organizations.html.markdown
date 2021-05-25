---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organizations"
sidebar_current: "docs-datasource-tfe-organizations"
description: |-
  Get information on Organizations.
---

# Data Source: tfe_organizations

Use this data source to get a list of Organizations and a map of their IDs.

## Example Usage

```hcl
data "tfe_organizations" "foo" {
}
```

## Argument Reference

No arguments are required. This retrieves the names and IDs of all the organizations.

## Attributes Reference

The following attributes are exported:

* `names` - A list of names of every organization.
* `ids` - A map of organization names and their IDs.
