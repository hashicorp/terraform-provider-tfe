# With an explicit expiration
# `expired_at` accepts an RFC3339/iso8601 timestamp and must be no more than 365 days in the future. You can use `time_rotating` to generate this dynamically:

resource "time_rotating" "example" {
  rotation_days = 30
}

resource "tfe_scim_token" "this" {
  description = "scim-token-30-day"
  expired_at  = time_rotating.example.rotation_rfc3339
  depends_on  = [tfe_scim_settings.this]
}
