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

# All trusted certificates, including legacy ones.
data "tfe_saml_idp_certificates" "all" {}

# Only certificates whose display name contains "us-east" (case-insensitive).
data "tfe_saml_idp_certificates" "us_east" {
  display_name_match = "us-east"
}

output "certificate_expirations" {
  value = {
    for c in data.tfe_saml_idp_certificates.all.certificates : c.id => {
      display_name = c.display_name
      expires_at   = c.expires_at
    }
  }
}
