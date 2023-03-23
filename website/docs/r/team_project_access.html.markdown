---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_project_access"
description: |-
  Associate a team to permissions on a project.
---

# tfe_team_project_access

Associate a team to permissions on a project.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "admin" {
  name         = "my-admin-team"
  organization = "my-org-name"
}

resource "tfe_project" "test" {
  name         = "myproject"
  organization = "my-org-name"
}

resource "tfe_team_project_access" "admin" {
  access       = "admin"
  team_id      = tfe_team.admin.id
  project_id   = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) ID of the team to add to the project.
* `project_id` - (Required) ID of the project to which the team will be added.
* `access` - (Required) Type of fixed access to grant. Valid values are `admin`, `maintain`, `write`, or `read`.

## Attributes Reference

* `id` The team project access ID.

## Import

Team project accesses can be imported; use the project team access ID as the import ID. For
example:

```shell
terraform import tfe_team_project_access.admin tprj-2pmtXpZa4YzVMTPi
```
