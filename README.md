<img alt="Terraform" src="https://www.datocms-assets.com/2885/1629941242-logo-terraform-main.svg" width="600px">


# Terraform Cloud/Enterprise Provider

The official Terraform provider for [Terraform Cloud/Enterprise](https://www.hashicorp.com/products/terraform).

As Terraform Enterprise is a self-hosted distribution of Terraform Cloud, this
provider supports both Cloud and Enterprise use cases. In all/most
documentation, the platform will always be stated as 'Terraform Enterprise' -
but a feature will be explicitly noted as only supported in one or the other, if
applicable (rare).

Note this provider is in beta and is subject to change (though it is generally
quite stable). We will indicate any breaking changes by releasing new versions.
Until the release of v1.0, any minor version changes will indicate possible
breaking changes. Patch version changes will be used for both bugfixes and
non-breaking changes.

- **Documentation:** https://registry.terraform.io/providers/hashicorp/tfe/latest/docs
- **Website**: https://registry.terraform.io/providers/hashicorp/tfe / https://www.terraform.io
- **Discuss forum**: https://discuss.hashicorp.com/c/terraform-providers

## Installation

Declare the provider in your configuration and `terraform init` will automatically fetch and install the provider for you from the [Terraform Registry](https://registry.terraform.io/):

```hcl
terraform {
  required_providers {
    tfe = {
      version = "~> 0.48.0"
    }
  }
}
```

For production use, you should constrain the acceptable provider versions via
configuration (as above), to ensure that new versions with breaking changes will
not be automatically installed by `terraform init` in the future. As this provider
is still at version zero, you should constrain the acceptable provider versions
on the minor version.

The above snippet using `required_providers` is for Terraform 0.13+; if you are using Terraform version 0.12, you can constrain by adding the version constraint to the `provider` block instead:

```hcl
provider "tfe" {
  version = "~> 0.48.0"
  ...
}
```

Since v0.24.0, this provider requires [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 0.12

For more information on provider installation and constraining provider versions, see the [Provider Requirements documentation](https://developer.hashicorp.com/terraform/language/providers/requirements).

## Usage

[Create a user or team API token in Terraform Cloud/Enterprise](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens), and use the token in the provider configuration block:

```hcl
provider "tfe" {
  hostname = var.hostname # Optional, for use with Terraform Enterprise. Defaults to app.terraform.io.
  token    = var.token
}

# Create an organization
resource "tfe_organization" "org" {
  # ...
}
```

There are several other ways to configure the authentication token, depending on
your use case. For other methods, see the [Authentication documentation](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs#authentication)

For more information on configuring providers in general, see the [Provider Configuration documentation](https://developer.hashicorp.com/terraform/language/providers/configuration).

# Development

We have developed some guidelines to help you learn more about compiling the provider, using it locally, and contributing suggested changes in the [contributing guide](https://hashicorp.github.io/terraform-provider-tfe/).
