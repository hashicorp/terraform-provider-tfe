# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_registry_provider" "example" {
  organization  = "my-org-name"
  registry_name = "public"
  namespace     = "hashicorp"
  name          = "aws"
}
