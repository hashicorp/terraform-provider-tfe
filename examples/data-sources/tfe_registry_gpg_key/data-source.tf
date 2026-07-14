# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_registry_gpg_key" "example" {
  organization = "my-org-name"
  id           = "13DFECCA3B58CE4A"
}
