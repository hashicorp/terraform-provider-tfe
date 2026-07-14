# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_variable_set" "test" {
  name         = "my-variable-set-name"
  organization = "my-org-name"
}
