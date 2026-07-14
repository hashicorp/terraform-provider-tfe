# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_ssh_key" "test" {
  name         = "my-ssh-key-name"
  organization = "my-org-name"
  key          = "private-ssh-key"
}
