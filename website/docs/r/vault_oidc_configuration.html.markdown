---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_vault_oidc_configuration"
description: |-
  Manages Vault OIDC configurations.
---

# tfe_vault_oidc_configuration

Defines a Vault OIDC configuration resource.

~> **NOTE:** This resource requires using the provider with HCP Terraform on the HCP Terraform Premium edition. Refer to [HCP Terraform pricing](https://www.hashicorp.com/en/pricing?product_intent=terraform&tab=terraform) for details.

## Example Usage

Basic usage:

```hcl
resource "tfe_vault_oidc_configuration" "example" {
  address           = "https://my-vault-cluster-public-vault-659decf3.b8298d98.z1.hashicorp.cloud:8200"
  role_name         = "vault-role-name"
  namespace         = "admin"
  auth_path         = "jwt-auth-path"
  encoded_cacert    = ""
  organization      = "my-org-name"
}
```


## Argument Reference

The following arguments are supported:

* `address` - (Required) The full address of your Vault instance.
* `role_name` - (Required) The name of a role in your Vault JWT auth path, with permission to encrypt and decrypt with a Transit secrets engine key.
* `namespace` - (Required) The namespace your JWT auth path is mounted in.
* `auth_path` - (Optional) 	The mounting path of JWT auth path of JWT auth. Defaults to `"jwt"`.
* `encoded_cacert` - (Optional) A base64 encoded certificate which can be used to authenticate your Vault certificate. Only needed for self-hosted Vault Enterprise instances with a self-signed certificate.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Vault OIDC configuration ID.

## Import
Vault OIDC configurations can be imported by ID.

Example:

```shell
terraform import tfe_vault_oidc_configuration.example voidc-AV61VxigiRvkkvPd
```
