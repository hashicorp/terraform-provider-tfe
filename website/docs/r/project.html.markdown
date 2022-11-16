---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project"
sidebar_current: "docs-resource-tfe-project"
description: |-
Manages projects.
---

# tfe_project

Provides a project resource.

~> **NOTE:** Projects functionality is currently in beta.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  organization = tfe_organization.test-organization.name
  name = "projectname"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the project.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The project ID.

## Import

Projects can be imported; use `<PROJECT ID>` as the import ID. For example:

```shell
terraform import tfe_project.test prj-niVoeESBXT8ZREhr
```
