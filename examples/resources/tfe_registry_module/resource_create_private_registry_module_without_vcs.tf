# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_module" "test-private-registry-module" {
  organization    = tfe_organization.test-organization.name
  module_provider = "my_provider"
  name            = "another_test_module"
  registry_name   = "private"
}
