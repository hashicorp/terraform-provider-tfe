# With write-only private key

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The Terraform Enterprise hostname."
  default     = "tfe.example.com"
}

variable "private_key" {
  type      = string
  ephemeral = true
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_idp_certificate" "primary" {
  display_name = "fooidp-us-east"
  cert         = file("${path.module}/fooidp-us-east.pem")
}

resource "tfe_saml_settings" "this" {
  slo_endpoint_url       = "https://example.com/slo_endpoint_url"
  sso_endpoint_url       = "https://example.com/sso_endpoint_url"
  private_key_wo         = var.private_key
  private_key_wo_version = 1

  depends_on = [tfe_saml_idp_certificate.primary]
}
