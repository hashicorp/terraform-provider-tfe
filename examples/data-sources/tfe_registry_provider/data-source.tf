# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# A private provider

data "tfe_registry_provider" "example" {
  organization = "my-org-name"
  name         = "my-provider"
}
