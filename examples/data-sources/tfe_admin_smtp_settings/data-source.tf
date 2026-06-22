provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_admin_smtp_settings" "foo" {
  provider = tfe.admin
}
