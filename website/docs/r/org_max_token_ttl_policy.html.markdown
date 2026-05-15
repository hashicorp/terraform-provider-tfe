---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_org_max_token_ttl_policy"
description: |-
  Manages the maximum time-to-live (TTL) policy for API tokens in an organization.
---

# tfe_org_max_token_ttl_policy

Manages the maximum time-to-live (TTL) policy for API tokens in an organization. This resource configures the maximum lifespans for organization, team, audit trail, and user tokens. Any tokens that exceed the configured limits will cease to work.

~> **Note:** To enable or disable the maximum TTL policy feature for an organization, use the `max_ttl_enabled` attribute on the `tfe_organization` resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform or an instance of Terraform Enterprise at least as recent as v2.0.1.

## Example Usage

```hcl
resource "tfe_organization" "test_org" {
  name            = "my-organization"
  email           = "admin@example.com"
  max_ttl_enabled = true  # Enable the max TTL policy feature
}

resource "tfe_org_max_token_ttl_policy" "token_ttl_policy" {
  organization              = tfe_organization.test_org.name
  org_token_max_ttl         = "0.5h"
  user_token_max_ttl        = "2.5y"
  team_token_max_ttl        = "3w"
  audit_trail_token_max_ttl = "6mo"
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `org_token_max_ttl` - (Optional) Maximum lifespan allowed for organization tokens to access the organization's resources. Defaults to two years (`2y`). Format: `<number><unit>` where unit is `h` (hours), `d` (days), `w` (weeks), `mo` (months), or `y` (years). Decimals are supported (e.g., `0.5h` for 30 minutes).
* `team_token_max_ttl` - (Optional) Maximum lifespan allowed for team tokens to access the organization's resources. Defaults to two years (`2y`). Format: `<number><unit>` where unit is `h` (hours), `d` (days), `w` (weeks), `mo` (months), or `y` (years). Decimals are supported (e.g., `0.5h` for 30 minutes).
* `audit_trail_token_max_ttl` - (Optional) Maximum lifespan allowed for audit trail tokens to access the organization's resources. Defaults to two years (`2y`). Format: `<number><unit>` where unit is `h` (hours), `d` (days), `w` (weeks), `mo` (months), or `y` (years). Decimals are supported (e.g., `0.5h` for 30 minutes).
* `user_token_max_ttl` - (Optional) Maximum lifespan allowed for user tokens to access the organization's resources. Defaults to two years (`2y`). Format: `<number><unit>` where unit is `h` (hours), `d` (days), `w` (weeks), `mo` (months), or `y` (years). Decimals are supported (e.g., `0.5h` for 30 minutes).

### TTL Format

All TTL attributes accept duration strings in the format `<number><unit>`:

| Unit | Description | Examples |
|------|-------------|----------|
| `h`  | Hours       | `1h`, `0.5h` (30 minutes), `12h`, `24h` |
| `d`  | Days        | `1d`, `2.5d`, `7d`, `30d` |
| `w`  | Weeks       | `1w`, `2w`, `4w` |
| `mo` | Months      | `1mo`, `3mo`, `6mo`, `12mo` |
| `y`  | Years       | `1y`, `2y` |

**Note:** Decimal values are supported for all units (e.g., `0.5h` = 30 minutes, `2.5d` = 2 days and 12 hours).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the token TTL policy (same as the organization name).
* `org_token_max_ttl_ms` - The computed maximum time-to-live for organization tokens, in milliseconds.
* `team_token_max_ttl_ms` - The computed maximum time-to-live for team tokens, in milliseconds.
* `audit_trail_token_max_ttl_ms` - The computed maximum time-to-live for audit trail tokens, in milliseconds.
* `user_token_max_ttl_ms` - The computed maximum time-to-live for user tokens, in milliseconds.

## Import

Token TTL policies can be imported using the organization name:

```shell
terraform import tfe_org_max_token_ttl_policy.example my-organization
```

~> **Note:** After import, the provider will convert the API's millisecond values to human-readable duration strings (e.g., `30d`, `6mo`, `2y`). If you prefer a different format (e.g., `720h` instead of `30d`), update your Terraform configuration accordingly. The provider will preserve your chosen format as long as it represents the same duration.
