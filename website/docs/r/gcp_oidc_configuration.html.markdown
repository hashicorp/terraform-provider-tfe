---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_gcp_oidc_configuration"
description: |-
  Manages GCP OIDC configurations.
---

# tfe_gcp_oidc_configuration

Defines a GCP OIDC configuration resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.

## Example Usage

Basic usage:

```hcl
resource "tfe_gcp_oidc_configuration" "example" {
  service_account_email     = "myemail@gmail.com"
  project_number            = "11111111"
  workload_provider_name    = "projects/1/locations/global/workloadIdentityPools/1/providers/1"
  organization              = "my-org-name"
}
```


## Argument Reference

The following arguments are supported:

* `service_account_email` - (Required) The email of your GCP service account, with permissions to encrypt and decrypt using a Cloud KMS key.
* `project_number` - (Required) The GCP Project containing the workload provider and service account.
* `workload_provider_name` - (Required) The fully qualified workload provider path. This should be in the format `projects/{project_number}/locations/global/workloadIdentityPools/{workload_identity_pool_id}/providers/{workload_identity_pool_provider_id}`.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The GCP OIDC configuration ID.

## Import
GCP OIDC configurations can be imported by ID.

Example:

```shell
terraform import tfe_gcp_oidc_configuration.example gcpoidc-PuXEeRoSaK3ENGj9
```
