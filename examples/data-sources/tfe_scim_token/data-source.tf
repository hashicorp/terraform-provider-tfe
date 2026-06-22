provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_scim_token" "foo" {
  provider = tfe.admin
  id       = "at-XXXXXXXXXXXXXXXX"
}
