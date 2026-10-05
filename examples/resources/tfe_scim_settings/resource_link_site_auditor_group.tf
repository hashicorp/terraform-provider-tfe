# The `site_auditor_group_scim_id` argument maps a SCIM group to the site auditor role, and works exactly like `site_admin_group_scim_id`. It also needs the two-apply workflow above, because the group must already exist in Terraform Enterprise. Requires Terraform Enterprise v2.1.0 or later.
# Linking a SCIM group to site auditor

variable "site_auditor_group_scim_id" {
  type        = string
  description = "SCIM ID of the group that should map to site auditor."
}

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
  provider_type    = "okta"
}

resource "tfe_scim_settings" "this" {
  site_auditor_group_scim_id = var.site_auditor_group_scim_id
  depends_on                 = [tfe_saml_settings.this]
}

# Clearing `site_auditor_group_scim_id` — setting it to `""` or removing it from your configuration — unlinks the group and revokes the site auditor role from every member of it.
