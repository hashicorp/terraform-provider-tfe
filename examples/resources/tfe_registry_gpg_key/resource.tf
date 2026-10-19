# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_registry_gpg_key" "example" {
  organization = "my-org-name"
  ascii_armor  = file("my-public-key.asc")
}
