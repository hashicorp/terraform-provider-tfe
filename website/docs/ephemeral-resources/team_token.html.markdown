---
layout: "tfe"
page_title: "Terraform Enterprise: Ephemeral: tfe_team_token"
description: |-
  Generates a new team token that is guaranteed not to be written to
  state.
---

# Ephemeral: tfe_team_token

Terraform ephemeral resource for managing a TFE team token. This
resource is used to generate a new team token that is guaranteed not to
be written to state. Since team tokens are singleton resources, using this ephemeral resource will replace any existing team token for a given team.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

### Generate a new team token:

This will invalidate any existing team token.

```hcl
resource "tfe_team" "example" {
  organization = "my-org-name"
  name = "my-team-name"
}

ephemeral "tfe_team_token" "example" {
  team_id = tfe_team.example.id
}
```

## Argument Reference

The following arguments are required:

* `team_id` - (Required) ID of the team.

The following arguments are optional:

* `expired_at` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339 
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and 
never expire.

This ephemeral resource exports the following attributes in addition to the arguments above:

* `token` - The generated token. This value is sensitive and will not be stored
  in state.
