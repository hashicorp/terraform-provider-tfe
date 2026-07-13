# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

variable "ssh_key" {
  type      = string
  ephemeral = true
}

resource "tfe_ssh_key" "test" {
  name           = "my-ssh-key-name"
  organization   = "my-org-name"
  key_wo         = var.ssh_key
  key_wo_version = 1
}
