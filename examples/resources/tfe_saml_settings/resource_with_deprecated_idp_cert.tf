# Deprecated usage with `idp_cert` (Terraform Enterprise releases earlier than v2.1.0)
#
# `idp_cert` is deprecated. It is still required on releases earlier than
# v2.1.0. After upgrading, move to `tfe_saml_idp_certificate` as shown above.

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

resource "tfe_saml_settings" "this" {
  idp_cert         = file("${path.module}/fooidp.pem")
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
}
