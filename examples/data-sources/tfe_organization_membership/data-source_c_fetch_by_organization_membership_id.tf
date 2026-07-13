# Fetch by organization membership ID

data "tfe_organization_membership" "test" {
  organization               = "my-org-name"
  organization_membership_id = "ou-xxxxxxxxxxx"
}
