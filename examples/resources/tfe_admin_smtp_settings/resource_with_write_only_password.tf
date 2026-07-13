# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

variable "admin_token" {
  description = "An admin access token"
}

variable "hostname" {
  description = "The HCP Terraform or Enterprise hostname."
  default     = "app.terraform.io"
}

variable "smtp_password" {
  type      = string
  ephemeral = true
}

provider "tfe" {
  hostname = var.hostname
  token    = var.admin_token
}

resource "tfe_admin_smtp_settings" "this" {
  host                = "smtp.example.com"
  port                = 587
  sender              = "noreply@example.com"
  auth                = "login"
  username            = "smtp_user"
  password_wo         = var.smtp_password
  password_wo_version = 1
}
