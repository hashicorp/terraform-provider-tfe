---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_policy_set_exclusion"
description: |-
  Add a policy set to an excluded project
---

# tfe_project_policy_set_exclusion

Adds and removes policy sets from an excluded project

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
* `project_id` - (Required) Excluded workspace ID to add the policy set to.

## Attributes Reference

* `id` - The ID of the policy set attachment. ID format: `<project-id>/<policy-set-id>`

## Import

Excluded Workspace Policy Sets can be imported; use `<PROJECT ID>/<POLICY SET ID>`. For example:

```shell
terraform import tfe_workspace_policy_set_exclusion.test 'prj-123456789/polset-123456789`
