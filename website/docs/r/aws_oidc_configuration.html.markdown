---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_aws_oidc_configuration"
description: |-
  Manages AWS OIDC configurations.
---

# tfe_aws_oidc_configuration

Defines an AWS OIDC configuration resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.

## Example Usage

Basic usage:

```hcl
resource "tfe_aws_oidc_configuration" "example" {
  role_arn      = "arn:aws:iam::111111111111:role/example-role-arn"
  organization  = "my-org-name"
}
```


## Argument Reference

The following arguments are supported:

* `role_arn` - (Required) The AWS ARN of your role..
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AWS OIDC configuration ID.

## Import
AWS OIDC configurations can be imported by ID.

Example:

```shell
terraform import tfe_aws_oidc_configuration.example awsoidc-DXmy3B2emVHysnbq
```
