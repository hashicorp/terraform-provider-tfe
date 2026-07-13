# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_provider_set" "my_provider_set" {
  name         = "example-provider-set"
  organization = "example-org"
}
