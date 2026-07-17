# With write-only private key

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

variable "private_key" {
  type      = string
  ephemeral = true
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_settings" "this" {
  idp_cert               = "foobarCertificate"
  slo_endpoint_url       = "https://example.com/slo_endpoint_url"
  sso_endpoint_url       = "https://example.com/sso_endpoint_url"
  private_key_wo         = var.private_key
  private_key_wo_version = 1
}
