# Linking a SCIM group to site admin (two-apply workflow) - Apply 2: link the site admin group

variable "site_admin_group_scim_id" {
  type        = string
  description = "SCIM ID of the group that should map to site admin."
}

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
  provider_type    = "okta"
}

resource "tfe_scim_settings" "this" {
  site_admin_group_scim_id = var.site_admin_group_scim_id
  depends_on               = [tfe_saml_settings.this]
}
