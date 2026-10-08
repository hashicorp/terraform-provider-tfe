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
  sso_endpoint_url = "https://example.com/sso"
  slo_endpoint_url = "https://example.com/slo"

  depends_on = [
    tfe_saml_idp_certificate.primary,
    tfe_saml_idp_certificate.failover,
  ]
}
