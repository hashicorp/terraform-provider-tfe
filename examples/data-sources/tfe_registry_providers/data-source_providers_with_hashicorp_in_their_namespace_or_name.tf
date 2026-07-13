# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

data "tfe_registry_providers" "hashicorp" {
  organization = "my-org-name"
  search       = "hashicorp"
}
