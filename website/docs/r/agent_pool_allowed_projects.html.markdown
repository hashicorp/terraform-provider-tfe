---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_pool_allowed_projects"
description: |-
  Manages allowed projects on agent pools
---

# tfe_agent_pool_allowed_projects

Adds and removes allowed projects on an agent pool.

~> **NOTE:** This resource requires using the provider with HCP Terraform and a HCP Terraform
for Business account.
[Learn more about HCP Terraform pricing here](https://www.hashicorp.com/products/terraform/pricing).

## Example Usage

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

// Ensure project and agent pool are create first
resource "tfe_project" "test-project" {
  name         = "my-project-name"
  organization = tfe_organization.test-organization.name
}

resource "tfe_agent_pool" "test-agent-pool" {
  name                = "my-agent-pool-name"
  organization        = tfe_organization.test-organization.name
  organization_scoped = false
}

// Ensure permissions are assigned second
resource "tfe_agent_pool_allowed_projects" "allowed" {
  agent_pool_id         = tfe_agent_pool.test-agent-pool.id
  allowed_project_ids   = [tfe_project.test-project.id]
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) The ID of the agent pool.
* `allowed_project_ids` - (Required) IDs of projects to be added as allowed projects on the agent pool.


## Import

A resource can be imported; use `<AGENT POOL ID>` as the import ID. For example:

```shell
terraform import tfe_agent_pool_allowed_projects.foobar apool-rW0KoLSlnuNb5adB
```
