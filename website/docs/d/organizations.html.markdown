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

No arguments are required. This retrieves the names and IDs of all the organizations readable by the provided token.

## Attributes Reference

The following attributes are exported:

* `names` - A list of names of every organization.
* `ids` - A map of organization names and their IDs.
* `admin` - A boolean field that determines  the list of organizations that should
  be retrieved. If it is true, then the [Admin Organizations
endpoint](https://www.terraform.io/docs/cloud/api/admin/organizations.html#list-all-organizations)
  will be used, otherwise if it is false or not included, then the [Organizations
endpoint](https://www.terraform.io/docs/cloud/api/organizations.html#list-organizations) will be used.
