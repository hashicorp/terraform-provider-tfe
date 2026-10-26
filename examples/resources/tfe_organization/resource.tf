# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Basic usage

resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@example.com"
}
