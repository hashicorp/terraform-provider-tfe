# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# All providers

data "tfe_registry_providers" "all" {
  organization = "my-org-name"
}
