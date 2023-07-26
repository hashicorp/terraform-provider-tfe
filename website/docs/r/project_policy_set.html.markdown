---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_policy_set"
description: |-
  Add a policy set to a project
---

# tfe_project_policy_set

Adds and removes policy sets from a project

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

resource "tfe_policy_set" "test" {
  name          = "my-policy-set"
  description   = "Some description."
  organization  = tfe_organization.test.name
}

resource "tfe_project_policy_set" "test" {
  policy_set_id = tfe_policy_set.test.id
  project_id    = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `project_id` - (Required) Project ID to add the policy set to.

## Attributes Reference

* `id` - The ID of the policy set attachment. ID format: `<project-id>_<policy-set-id>`

## Import

Project Policy Sets can be imported; use `<ORGANIZATION>/<PROJECT ID>/<POLICY SET NAME>`. For example:

```shell
terraform import tfe_project_policy_set.test 'my-org-name/project/policy-set-name'
```
