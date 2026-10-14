# Recommended usage for SAML Settings (Terraform Enterprise v2.1.0 or later)
#
# Manage IdP certificates with `tfe_saml_idp_certificate` and leave `idp_cert`
# unset. Terraform Enterprise does not allow enabling SAML without a trusted
# certificate, so `depends_on` creates the certificates before SAML is enabled
# and disables SAML before the certificates are destroyed.

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The Terraform Enterprise hostname."
  default     = "tfe.example.com"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_idp_certificate" "primary" {
  display_name = "fooidp-us-east"
  cert         = file("${path.module}/fooidp-us-east.pem")
}

resource "tfe_saml_idp_certificate" "failover" {
  display_name = "fooidp-eu-west"
  cert         = file("${path.module}/fooidp-eu-west.pem")
}

resource "tfe_saml_settings" "this" {
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"

  depends_on = [
    tfe_saml_idp_certificate.primary,
    tfe_saml_idp_certificate.failover,
  ]
}
