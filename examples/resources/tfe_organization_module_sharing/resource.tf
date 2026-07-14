# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization_module_sharing" "test" {
  organization     = "my-org-name"
  module_consumers = ["my-org-name-2", "my-org-name-3"]
}
