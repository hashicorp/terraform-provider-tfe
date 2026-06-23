data "tfe_current_user" "current" {}

resource "tfe_organization_membership" "owner" {
  organization = "my-org"
  email        = data.tfe_current_user.current.email
}
