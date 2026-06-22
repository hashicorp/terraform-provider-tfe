provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_admin_smtp_settings" "this" {
  host   = "smtp.example.com"
  port   = 25
  sender = "noreply@example.com"
  auth   = "none"
}
