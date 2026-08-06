# Basic usage

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "user@example.com"
}
