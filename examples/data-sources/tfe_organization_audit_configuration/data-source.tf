# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_organization_audit_configuration" "example" {
  organization = "my-org-name"
}
