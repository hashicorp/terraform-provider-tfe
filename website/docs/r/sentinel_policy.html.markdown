---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_sentinel_policy"
description: |-
  Manages Sentinel policies.
---

# tfe_sentinel_policy

Sentinel Policy as Code is an embedded policy as code framework integrated
with Terraform Enterprise.

Policies are configured on a per-organization level and are organized and
grouped into policy sets, which define the workspaces on which policies are
enforced during runs.

## Example Usage

Basic usage:

```hcl
resource "tfe_sentinel_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy.
* `description` - (Optional) A description of the policy's purpose.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `policy` - (Required) The actual policy itself.
* `enforce_mode` - (Optional) The enforcement level of the policy. Valid
  values are `advisory`, `hard-mandatory` and `soft-mandatory`. Defaults
  to `soft-mandatory`.

## Attributes Reference

* `id` - The ID of the policy.

## Import

Sentinel policies can be imported; use `<ORGANIZATION NAME>/<POLICY ID>` as the
import ID. For example:

```shell
terraform import tfe_sentinel_policy.test my-org-name/pol-wAs3zYmWAhYK7peR
```
