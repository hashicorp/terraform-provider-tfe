provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_scim_group" "site_admins" {
  provider = tfe.admin
  name     = "tfe-site-admins"
}

resource "tfe_scim_settings" "this" {
  provider                 = tfe.admin
  site_admin_group_scim_id = data.tfe_scim_group.site_admins.id
}
