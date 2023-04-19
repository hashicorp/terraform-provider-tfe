---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_variable_set"
description: |-
  Add a variable set to a project
---

# tfe_project_variable_set

Adds and removes variable sets from a project

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  name         = "my-project-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable_set" "test" {
  name         = "Test Varset"
  description  = "Some description."
  organization = tfe_organization.test.name
}

resource "tfe_project_variable_set" "test" {
  variable_set_id = tfe_variable_set.test.id
  project_id      = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `variable_set_id` - (Required) Name of the variable set to add.
* `project_id` - (Required) Project ID to add the variable set to.

## Attributes Reference

* `id` - The ID of the variable set attachment. ID format: `<project-id>_<variable-set-id>`

## Import

Project Variable Sets can be imported; use `<ORGANIZATION>/<PROJECT ID>/<VARIABLE SET NAME>`. For example:

```shell
terraform import tfe_project_variable_set.test 'my-org-name/prj-F1NpdVBuCF3xc5Rp/Test Varset'
```
