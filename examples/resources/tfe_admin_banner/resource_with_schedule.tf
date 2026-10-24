# Schedule a banner to appear and expire automatically.
# scheduled_publish_at, scheduled_expire_at and timezone must be set together.
# The banner remains in Terraform state after it expires; remove the resource
# from your configuration (or run `terraform destroy`) to deactivate it.

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The Terraform Enterprise hostname."
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_admin_banner" "maintenance_window" {
  title                = "Maintenance window"
  body                 = "Planned maintenance from 02:00 to 04:00 UTC."
  style                = "warning"
  audience             = "authenticated_only"
  scheduled_publish_at = "2026-12-06T00:00:00Z"
  scheduled_expire_at  = "2026-12-06T04:00:00Z"
  timezone             = "UTC"
}
