---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_settings_cost_estimation"
sidebar_current: "docs-resource-tfe-settings-cost-estimation"
description: |-
  Manage the cost estimation settings of a Terraform Enterprise installation.
---

# tfe_settings_cost_estimation

Manage the [cost estimation settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-cost-estimation-settings) of a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_settings_cost_estimation" "settings" {
  enabled = true

  aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
  aws_secret_key        = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  gcp_credentials       = "{\"private_key\":\"-----BEGIN PRIVATE KEY-----\\n....=\\n-----END PRIVATE KEY-----\",\"private_key_id\":\"some_id\",...}"
  azure_client_id       = "9b516fe8-415s-9119-bab0-EXAMPLEID1"
  azure_client_secret   = "9b516fe8-415s-9119-bab0-EXAMPLESEC1"
  azure_subscription_id = "9b516fe8-415s-9119-bab0-EXAMPLEID2"
  azure_tenant_id       = "9b516fe8-415s-9119-bab0-EXAMPLEID3"
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Optional) Allows organizations to opt-in to the Cost Estimation feature. Default to `false`.
* `aws_access_key_id` - (Optional) An AWS Access Key ID that the Cost Estimation feature will use to authorize to AWS's Pricing API.
* `aws_secret_key` - (Optional) An AWS Secret Key that the Cost Estimation feature will use to authorize to AWS's Pricing API.
* `gcp_credentials` - (Optional) A JSON string containing GCP credentials that the Cost Estimation feature will use to authorize to the Google Cloud Platform's Pricing API.
* `azure_client_id` - (Optional) An Azure Client ID that the Cost Estimation feature will use to authorize to Azure's RateCard API.
* `azure_client_secret` - (Optional) An Azure Client Secret that the Cost Estimation feature will use to authorize to Azure's RateCard API.
* `azure_subscription_id` - (Optional) An Azure Subscription ID that the Cost Estimation feature will use to authorize to Azure's RateCard API.
* `azure_tenant_id` - (Optional) An Azure Tenant ID that the Cost Estimation feature will use to authorize to Azure's RateCard API.
