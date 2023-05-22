---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_token"
description: |-
  Manages agent tokens
---

# tfe_agent_token

Each agent pool has its own set of tokens which are not shared across pools.
These tokens allow agents to communicate securely with Terraform Cloud.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "test-agent-pool" {
  name         = "my-agent-pool-name"
  organization = tfe_organization.test-organization.id
}

resource "tfe_agent_token" "test-agent-token" {
  agent_pool_id = tfe_agent_pool.test-agent-pool.id
  description   = "my-agent-token-name"
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) ID of the agent pool.
* `description` - (Required) Description of the agent token.

## Attributes Reference

* `id` - The ID of the agent token.
* `description` - The description of agent token.
* `token` - The generated token.
