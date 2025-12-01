---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_stack_variable_set"
description: |-
  Add a variable set to a stack
---

# tfe_stack_variable_set

Adds and removes a stack from a variable set's scope.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_stack" "test" {
  organization  = tfe_organization.test.id
  name          = "my-stack-name"
}

resource "tfe_variable_set" "test" {
  name         = "Test Varset"
  description  = "Some description."
  organization = tfe_organization.test.id
}

resource "tfe_stack_variable_set" "test" {
  stack_id        = tfe_stack.test.id
  variable_set_id = tfe_variable_set.test.id
}
```

## Argument Reference

The following arguments are supported:

* `variable_set_id` - (Required) ID of the variable set to add to the stack.
* `stack_id` - (Required) ID of the stack to add the variable set to.

## Attributes Reference

* `id` - The ID of the stack variable set association in the format `stack/varset`.

## Import

Stack variable sets can be imported using the ID in the format `stack/varset`.

Example:

```shell
terraform import tfe_stack_variable_set.test st-abcdefgh/varset-ijklmnop
```
