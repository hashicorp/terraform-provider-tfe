# Look up a single SCIM group by its exact name and reference its ID

variable "token" {
  description = "An access token"
}

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

provider "tfe" {
  hostname = var.hostname
  token    = var.token
}

provider "tfe" {
  alias    = "admin"
  hostname = var.hostname
  token    = var.admin_token
}

data "tfe_scim_group" "admins" {
  provider = tfe.admin
  name     = "platform-admins"
}

output "admin_group_id" {
  value = data.tfe_scim_group.admins.id
}
