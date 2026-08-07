# Basic usage

data "tfe_current_user" "current" {}

output "email" {
  value = data.tfe_current_user.current.email
}
