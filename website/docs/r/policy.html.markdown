---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_policy"
description: |-
  Manages policies.
---

# tfe_policy

Policies are rules enforced on Terraform runs. You can use policies to validate that the Terraform plan complies with security rules and best practices.
Two policy-as-code frameworks are integrated with Terraform Enterprise: Sentinel and Open Policy Agent (OPA).

Policies are configured on a per-organization level and are organized and
grouped into policy sets, which define the workspaces on which policies are
enforced during runs.

## Example Usage

Basic usage for Sentinel:

```hcl
resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "sentinel"
  policy       = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}
```

Basic usage for Open Policy Agent(OPA):

```hcl
resource "tfe_policy" "test" {
  name         = "my-policy-name"
  description  = "This policy always passes"
  organization = "my-org-name"
  kind         = "opa"
  policy       = "package example rule[\"not allowed\"] { false }"
  query        = "data.example.rule"
  enforce_mode = "mandatory"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy.
* `description` - (Optional) A description of the policy's purpose.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `kind` - (Optional) The policy-as-code framework associated with the policy.
   Defaults to `sentinel` if not provided. Valid values are `sentinel` and `opa`.
* `query` - (Optional) The OPA query to identify a specific policy rule that
   needs to run within your Rego code. Required for all OPA policies.
* `policy` - (Required) The actual policy itself.
* `enforce_mode` - (Optional) The enforcement level of the policy. Valid
  values for Sentinel are `advisory`, `hard-mandatory` and `soft-mandatory`. Defaults
  to `soft-mandatory`. Valid values for OPA are `advisory` and `mandatory`. Defaults
  to `advisory`.

## Attributes Reference

* `id` - The ID of the policy.

## Import

Policies can be imported; use `<ORGANIZATION NAME>/<POLICY ID>` as the
import ID. For example:

```shell
terraform import tfe_policy.test my-org-name/pol-wAs3zYmWAhYK7peR
```
