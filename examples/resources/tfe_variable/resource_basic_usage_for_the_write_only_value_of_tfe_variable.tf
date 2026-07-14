# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.name
}

resource "tfe_variable" "test" {
  key              = "my_key_name"
  value_wo         = var.session_token
  value_wo_version = 1
  category         = "terraform"
  workspace_id     = tfe_workspace.test.id
  description      = "a useful description"
}
