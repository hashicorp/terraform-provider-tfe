# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# All private providers

data "tfe_registry_providers" "private" {
  organization  = "my-org-name"
  registry_name = "private"
}
