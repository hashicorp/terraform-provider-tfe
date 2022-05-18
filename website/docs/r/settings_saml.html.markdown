---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_settings_saml"
sidebar_current: "docs-resource-tfe-settings-saml"
description: |-
  Manage the SAML settings of a Terraform Enterprise installation.
---

# tfe_settings_saml

Manage the [SAML settings](https://www.terraform.io/cloud-docs/api-docs/admin/settings#list-saml-settings) of a Terraform Enterprise installation.

## Example Usage

Basic usage:

```hcl
resource "tfe_settings_saml" "settings" {
  enabled = true
  debug   = false

  idp_cert                      = "NEW-CERTIFICATE"
  slo_endpoint_url              = "https://example.com/slo"
  sso_endpoint_url              = "https://example.com/sso"
  attr_username                 = "Username"
  attr_groups                   = "MemberOf"
  attr_site_admin               = "SiteAdmin"
  site_admin_role               = "site-admins"
  sso_api_token_session_timeout = 1209600
}
```

## Argument Reference

The following arguments are supported:

* `enabled` - (Optional) Allows SAML to be used. If true, all remaining attributes must have valid values. Default to `false`.
* `debug` - (Optional) Enables a SAML debug dialog that allows an admin to see the SAMLResponse XML and processed values during login. Default to `false`.
* `idp_cert` - (Optional) Identity Provider Certificate specifies the PEM encoded X.509 Certificate as provided by the IdP configuration.
* `slo_endpoint_url` - (Optional) Single Log Out URL specifies the HTTPS endpoint on your IdP for single logout requests. This value is provided by the IdP configuration.
* `sso_endpoint_url` - (Optional) Single Sign On URL specifies the HTTPS endpoint on your IdP for single sign-on requests. This value is provided by the IdP configuration.
* `attr_username` - (Optional) Username Attribute Name specifies the name of the SAML attribute that determines the user's username. Default to `"Username"`.
* `attr_groups` - (Optional) Team Attribute Name specifies the name of the SAML attribute that determines team membership. Default to `"MemberOf"`.
* `attr_site_admin` - (Optional) Specifies the role for site admin access. Overrides the "Site Admin Role" method. Default to `"SiteAdmin"`.
* `site_admin_role` - (Optional) Specifies the role for site admin access, provided in the list of roles sent in the Team Attribute Name attribute. Default to `"site-admins"`.
* `sso_api_token_session_timeout` - (Optional) Specifies the Single Sign On session timeout in seconds. Default to 1209600 (14 days).
