# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_github_app_installation" "gha_installation" {
  installation_id = 12345678
}
