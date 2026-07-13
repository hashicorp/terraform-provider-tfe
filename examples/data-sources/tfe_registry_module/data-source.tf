# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_registry_module" "example" {
  organization    = "my-organization"
  name            = "no-code-ssm"
  module_provider = "aws"
}
