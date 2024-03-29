---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_saml_settings"
description: |-
  Get information on SAML Settings.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_saml_settings

Use this data source to get information about SAML Settings. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.


## Example Usage

Basic usage:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_saml_settings" "foo" {
  provider = tfe.admin
}
```

## Argument Reference

No arguments are required for this data source.

## Attributes Reference

The following attributes are exported:

* `Id` - It is always `Saml`.
* `Enabled` - Whether SAML single sign-on is enabled.
* `Debug` - Whether debug mode is enabled, which means that the SAMLResponse XML will be displayed on the login page.
* `TeamManagementEnabled` - Whether Terraform Enterprise is set to manage team membership.
* `AuthnRequestsSigned` - Whether `<samlp:AuthnRequest>` messages are signed.
* `WantAssertionsSigned` - Whether `<saml:Assertion>` elements are signed.
* `IdpCert` - PEM encoded X.509 Certificate as provided by the IdP configuration.
* `OldIdpCert` - Previous version of the PEM encoded X.509 Certificate as provided by the IdP configuration.
* `SloEndpointUrl` - Single Log Out URL.
* `SsoEndpointUrl` - Single Sign On URL.
* `AttrUsername` - Name of the SAML attribute that determines the user's username.
* `AttrGroups` - Name of the SAML attribute that determines team membership.
* `AttrSiteAdmin` - Site admin access role.
* `SiteAdminRole` - Site admin access role.
* `SsoApiTokenSessionTimeout` - Single Sign On session timeout in seconds.
* `AcsConsumerUrl` - ACS Consumer (Recipient) URL.
* `MetadataUrl` - Metadata (Audience) URL.
* `Certificate` - Request and assertion signing certificate.
* `Certificate` - Request and assertion signing certificate.
* `PrivateKey` - The private key used for request and assertion signing.
* `SignatureSigningMethod` - Signature Signing Method.
* `SignatureDigestMethod` - Signature Digest Method.

<!-- cache-key: cdktf-0.17.0-pre.15 input-2995e79c51b29afd8d8b89c5d98dae47701709e1502f993d5be8b2b681de4895 -->