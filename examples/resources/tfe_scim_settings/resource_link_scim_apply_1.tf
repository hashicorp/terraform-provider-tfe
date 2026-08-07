# The `site_admin_group_scim_id` argument links a SCIM group to the site admin role. The group must already exist in Terraform Enterprise, but groups are only created by your IdP after SCIM provisioning is enabled. This requires a two-apply workflow:
# Linking a SCIM group to site admin - Apply 1: enable SCIM

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
  provider_type    = "okta"
}

resource "tfe_scim_settings" "this" {
  depends_on = [tfe_saml_settings.this]
}
