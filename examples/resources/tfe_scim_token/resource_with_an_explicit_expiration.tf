# With an explicit expiration

resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_scim_token" "this" {
  description = "scim-token-30-day"
  expired_at  = time_rotating.example.rotation_rfc3339
  depends_on  = [tfe_scim_settings.this]
}
