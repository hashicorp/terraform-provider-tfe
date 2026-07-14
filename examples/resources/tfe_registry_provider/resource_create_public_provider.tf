# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_organization" "example" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_registry_provider" "example" {
  organization = tfe_organization.example.name

  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
