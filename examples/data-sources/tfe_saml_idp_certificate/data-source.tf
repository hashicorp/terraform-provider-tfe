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

data "tfe_saml_idp_certificate" "primary" {
  id = "idpc-AbC123XyZ456"
}

output "primary_expires_at" {
  value = data.tfe_saml_idp_certificate.primary.expires_at
}
