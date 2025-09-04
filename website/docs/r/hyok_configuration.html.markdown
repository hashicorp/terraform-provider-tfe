---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_hyok_configuration"
description: |-
  Manages HYOK configurations.
---

# tfe_hyok_configuration

Defines a HYOK configuration resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.

## Example Usage

Basic usage:

```hcl
resource "tfe_hyok_configuration" "gcp_example" {
  organization              = "my-hyok-org"
  name                      = "my-key-name"
  kek_id                    = "key1"
  agent_pool_id             = "apool-MFtsuFxHkC9pCRgB"
  gcp_oidc_configuration_id = "gcpoidc-PuXEeRoSaK3ENGj9"

  kms_options {
    key_location  = "global"
    key_ring_id   = "example-key-ring"
  }
}
```


## Argument Reference

The following arguments are supported:
* `name` - (Required) Label for the HYOK configuration to be used within HCP Terraform.
* `kek_id` - (Required) Refers to the name of your key encryption key stored in your key management service.
* `agent_pool_id` - (Required) The ID of the agent-pool to associate with the HYOK configuration.
* `vault_oidc_configuration_id` - (Optional) The ID of the TFE Vault OIDC configuration. If this is set, no other OIDC configuration IDs should be set.
* `aws_oidc_configuration_id` - (Optional) The ID of the TFE AWS OIDC configuration. If this is set, no other OIDC configuration IDs should be set.
* `gcp_oidc_configuration_id` - (Optional) The ID of the TFE GCP OIDC configuration. If this is set, no other OIDC configuration IDs should be set.
* `azure_oidc_configuration_id` - (Optional) The ID of the TFE Azure OIDC configuration. If this is set, no other OIDC configuration IDs should be set.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

The `kms_options` block is optional, and is used to specify additional fields for some key management services. Supported arguments are:
* `key_region` - (Optional) The AWS region where your key is located.
* `key_location` - (Optional) The location in which the GCP key ring exists.
* `key_ring_id` - (Optional) The root resource for Google Cloud KMS keys and key versions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The HYOK configuration ID.

## Import
HYOK configurations can be imported by ID.

Example:

```shell
terraform import tfe_hyok_configuration.gcp_example hyokc-XqYizSPQmeiG1aHJ
```
