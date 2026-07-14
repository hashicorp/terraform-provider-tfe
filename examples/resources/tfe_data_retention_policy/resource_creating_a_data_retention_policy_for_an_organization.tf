# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "test-organization" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_data_retention_policy" "foobar" {
  organization = tfe_organization.test-organization.name

  delete_older_than {
    days = 1138
  }
}
