---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_token"
description: |-
  Generates an ephemeral agent token.
---

# tfe_agent_token

Generates a new agent token as an ephemeral value.

## Example Usage

Basic usage:

```hcl
ephemeral "tfe_agent_token" "this" {
  agent_pool_id = tfe_agent_pool.foobar.id
  description   = "my description"
}
```

## Argument Reference

The following arguments are supported:

* `agent_pool_id` - (Required) Id for the Agent Pool.
* `description` - (Required) A brief description about the Agent Pool.

## Example Usage

```hcl
resource "tfe_agent_pool" "foobar" {
  name         = "agent-pool-test"
  organization = my-org-name"
}

ephemeral "tfe_agent_token" "this" {
  agent_pool_id = tfe_agent_pool.foobar.id
  description   = "my description"
}

output "my-agent-token" {
  value       = ephemeral.tfe_agent_token.this.token
  description = "Token for tfe agent."
  ephemeral   = true
}
```

## Attributes Reference

* `token` - The generated token.

