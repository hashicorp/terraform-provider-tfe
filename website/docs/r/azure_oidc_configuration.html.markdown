---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_azure_oidc_configuration"
description: |-
  Manages Azure OIDC configurations.
---

# tfe_azure_oidc_configuration

Defines an Azure OIDC configuration resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.

## Example Usage

Basic usage:

```hcl
resource "tfe_azure_oidc_configuration" "example" {
  client_id         = "application-id1"
  subscription_id   = "subscription-id1"
  tenant_id         = "tenant-id1"
  organization      = "my-org-name"
}
```


## Argument Reference

The following arguments are supported:

* `client_id` - (Required) The Client (or Application) ID of your Entra ID application.
* `subscription_id` - (Required) The ID of your Azure subscription.
* `tenant_id` - (Required) The Tenant (or Directory) ID of your Entra ID application.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Azure OIDC configuration ID.

## Import
Azure OIDC configurations can be imported by ID.

Example:

```shell
terraform import tfe_azure_oidc_configuration.example azoidc-8DCgwEV2GbMcQjk8
```
