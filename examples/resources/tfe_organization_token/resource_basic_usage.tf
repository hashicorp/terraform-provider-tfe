resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_organization_token" "test" {
  organization = data.tfe_organization.org.name
  expired_at   = time_rotating.example.rotation_rfc3339
}
