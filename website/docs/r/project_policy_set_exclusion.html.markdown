---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_policy_set_exclusion"
description: |-
  Exclude a project from a policy set
---

# tfe_project_policy_set_exclusion

Adds and removes project exclusions from a policy set.

-> **Note:** `tfe_policy_set` has an argument `global` that should be `true` to use this resource.

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
  global        = true
}

resource "tfe_project_policy_set_exclusion" "test" {
  policy_set_id = tfe_policy_set.test.id
  project_id    = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `policy_set_id` - (Required) ID of the policy set.
* `project_id` - (Required)  The Project ID where HCP Terraform excludes the specified policy set.

## Attributes Reference

* `id` - The ID of the project policy set exclusion. ID format: `<project-id>/<policy-set-id>`

## Import

Excluded Project Policy Sets can be imported; use `<PROJECT ID>/<POLICY SET ID>`. For example:

```shell
terraform import tfe_project_policy_set_exclusion.test 'prj-123456789/polset-123456789`
