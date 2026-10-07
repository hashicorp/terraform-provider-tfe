# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Create private provider

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  name = "my-provider"
}
