---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_agent_token"
description: |-
  Generates an ephemeral agent token.
---

# tfe_agent_token

Generates an ephemeral agent token for use during a Terraform run.

Ephemeral agent pool tokens are only valid within the context of a single run, and
are not stored in Terraform state.

Ephemeral resources are provisioned during the plan phase of a run as well as
the apply phase.

If you need the agent token to remain valid for long-lived use, consider using the
`tfe_agent_token` managed resource instead.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

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
  organization = "my-org-name"
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

