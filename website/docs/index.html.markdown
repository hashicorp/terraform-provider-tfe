---
layout: "tfe"
page_title: "Provider: Terraform Enterprise"
sidebar_current: "docs-tfe-index"
description: |-
  The Terraform Enterprise provider is used to interact with the many resources supported by (Private) Terraform Enterprise. The provider needs to be configured with the proper credentials before it can be used.
---

# Terraform Enterprise Provider

The Terraform Enterprise provider is used to interact with the many resources
supported by (Private) [Terraform Enterprise](https://www.hashicorp.com/products/terraform).
The provider needs to be configured with the proper credentials before it can
be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the Terraform Enterprise Provider
provider "tfe" {
  hostname = "${var.hostname}"
  token    = "${var.token}"
}

# Create an organization
resource "tfe_organization" "org" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `hostname` - (Optional) The Terraform Enterprise hostname to connect to.
  Defaults to `app.terraform.io`.
* `token` - (Optional) The token used to authenticate with Terraform Enterprise.
	We recommend omitting the token which can be set as `credentials` in the
  [CLI config file](/docs/commands/cli-config.html#credentials).
