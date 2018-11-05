---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_sentinel_policy"
sidebar_current: "docs-resource-tfe-sentinel-policy"
description: |-
  Sentinel Policy as Code is an embedded policy as code framework integrated with Terraform Enterprise.
---

# tfe_sentinel_policy

Sentinel Policy as Code is an embedded policy as code framework integrated
with Terraform Enterprise.

Policies are configured on a per-organization level, and are enforced on
all of an organization's workspaces during runs. Each plan's changes are
validated against the policy prior to the apply step.

## Example Usage

Basic usage:

```hcl
resource "tfe_sentinel_policy" "test" {
  name = "my-policy-name"
  organization = "my-org-name"
  policy = "main = rule { true }"
  enforce_mode = "hard-mandatory"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy.
* `organization` - (Required) Name of the organization.
* `policy` - (Required) The actual policy itself.
* `enforce_mode` - (Required) The enforcement level of the policy. Valid
  values are `advisory`, `hard-mandatory` and `soft-mandatory`. Defaults
  to `soft-mandatory`.

## Attributes Reference

* `id` - The ID of the policy.

## Import

Sentinel policies can be imported with an ID of `<ORGANIZATION NAME>/<RESOURCE ID>`. For example:

```shell
terraform import tfe_sentinel_policy.test my-org-name/pol-wAs3zYmWAhYK7peR
```
