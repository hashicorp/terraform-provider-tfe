---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_saml_settings"
description: |-
  Manages SAML Settings.
---

# tfe_saml_settings

Use this resource to create, update and destroy SAML Settings. It applies only to Terraform Enterprise and requires admin token configuration. See example usage for incorporating an admin token in your provider config.

## Example Usage

Basic usage for SAML Settings:

```hcl
provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
 }
```

## Argument Reference

The following arguments are supported:

* `idp_cert` - (Required) Identity Provider Certificate specifies the PEM encoded X.509 Certificate as provided by the IdP configuration.
* `slo_endpoint_url` - (Required) Single Log Out URL specifies the HTTPS endpoint on your IdP for single logout requests. This value is provided by the IdP configuration.
* `sso_endpoint_url` - (Required) Single Sign On URL specifies the HTTPS endpoint on your IdP for single sign-on requests. This value is provided by the IdP configuration.
* `debug` - (Optional) When sign-on fails and this is enabled, the SAMLResponse XML will be displayed on the login page.
* `authn_requests_signed` - (Optional) Whether to ensure that `<samlp:AuthnRequest>` messages are signed.
* `want_assertions_signed` - (Optional) Whether to ensure that `<samlp:Assertion>` elements are signed.
* `team_management_enabled` - (Optional) Set it to false if you would rather use Terraform Enterprise to manage team membership.
* `attr_username` - (Optional) Username Attribute Name specifies the name of the SAML attribute that determines the user's username.
* `attr_site_admin` - (Optional) Specifies the role for site admin access. Overrides the `Site Admin Role` method.
* `attr_groups` - (Optional) Team Attribute Name specifies the name of the SAML attribute that determines team membership.
* `site_admin_role` - (Optional) Specifies the role for site admin access, provided in the list of roles sent in the Team Attribute Name attribute.
* `sso_api_token_session_timeout` - (Optional) Specifies the Single Sign On session timeout in seconds. Defaults to 14 days.
* `certificate` - (Optional) The certificate used for request and assertion signing.
* `private_key` - (Optional) The private key used for request and assertion signing.
* `signature_signing_method` - (Optional) Signature Signing Method. Must be either `SHA1` or `SHA256`. Defaults to `SHA256`.
* `signature_digest_method` - (Optional) Signature Digest Method. Must be either `SHA1` or `SHA256`. Defaults to `SHA256`.

## Attributes Reference

* `id` - The ID of the SAML Settings. Always `saml`.
* `acs_consumer_url` - ACS Consumer (Recipient) URL.
* `metadata_url` - Metadata (Audience) URL.
* `old_idp_cert` - Value of the old IDP Certificate.

## Import

SAML Settings can be imported.

```shell
terraform import tfe_saml_settings.this saml
```
