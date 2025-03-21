---
layout: "tfe"
page_title: "Terraform Enterprise: Ephemeral: tfe_outputs"
description: |-
  Get output values from another organization/workspace without writing
  sensitive data to state.
---

# Ephemeral: tfe_organization_token

Terraform ephemeral resource for managing a TFE organization token. This
resource is used to generate a new organization token that is guaranteed not to
be written to state. Since organization tokens are singleton resources, using this ephemeral resource will replace any existing organization token, including those managed by `tfe_organization_token`.

This ephemeral resource is used to retrieve the state outputs for a given workspace.
It enables output values in one Terraform configuration to be used in another.
The retrieved output values are guaranteed not to be written to state.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

---


~> **NOTE:** The `values` attribute is preemptively marked [sensitive](https://developer.hashicorp.com/terraform/language/values/outputs#sensitive-suppressing-values-in-cli-output) and is only populated after a run completes on the associated workspace. Use the `nonsensitive_values` attribute to access the subset of the outputs
that are known to be non-sensitive.

## Example Usage

Using the `tfe_outputs` ephemeral resource, the outputs `vault_role_id` and `vault_secret_id` can be used to configure a vault provider instance as seen below:

In the example below, assume we have outputs defined in a `my-org/my-workspace`:

```hcl
ephemeral "tfe_outputs" "foo" {
  organization = "my-org"
  workspace = "my-workspace"
}

provider "vault" {
  auth_login {
    path = "auth/approle/login"

    parameters = {
      role_id   = ephemeral_tfe_outputs.foo.values.vault_role_id
      secret_id = ephemeral_tfe_outputs.foo.values.vault_secret_id
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `organization` - (Required) The name of the organization.
* `workspace` - (Required) The name of the workspace.

## Attributes Reference

The following attributes are exported:

* `values` - The current output values for the specified workspace.
* `nonsensitive_values` - The current non-sensitive output values for the specified workspace, this is a subset of all output values.
