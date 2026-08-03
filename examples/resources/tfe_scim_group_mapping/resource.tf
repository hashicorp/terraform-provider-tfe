# The SCIM group you map to must already exist in Terraform Enterprise. Groups are created by your IdP after SCIM provisioning is enabled, so this typically follows a workflow where SCIM is enabled first, the IdP pushes the group, and then the mapping is created. Use the tfe_scim_group data source to look up the group's ID by name:

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_saml_settings" "this" {
  idp_cert         = "foobarCertificate"
  slo_endpoint_url = "https://example.com/slo_endpoint_url"
  sso_endpoint_url = "https://example.com/sso_endpoint_url"
  provider_type    = "okta"
}

resource "tfe_scim_settings" "this" {
  depends_on = [tfe_saml_settings.this]
}

data "tfe_organization" "this" {
  name = "my-org-name"
}

resource "tfe_team" "engineering" {
  name         = "engineering"
  organization = data.tfe_organization.this.name
}

# Look up the SCIM group provisioned by your IdP.
data "tfe_scim_group" "engineering" {
  name       = "engineering-scim-group"
  depends_on = [tfe_scim_settings.this]
}

resource "tfe_scim_group_mapping" "engineering" {
  team_id       = tfe_team.engineering.id
  scim_group_id = data.tfe_scim_group.engineering.id
}
