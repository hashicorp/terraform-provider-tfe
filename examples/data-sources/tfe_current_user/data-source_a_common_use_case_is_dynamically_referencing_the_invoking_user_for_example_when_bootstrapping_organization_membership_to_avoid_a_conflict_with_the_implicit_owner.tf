# A common use case is dynamically referencing the invoking user, for example when bootstrapping organization membership to avoid a conflict with the implicit owner

data "tfe_current_user" "current" {}

resource "tfe_organization_membership" "owner" {
  organization = "my-org"
  email        = data.tfe_current_user.current.email
}
