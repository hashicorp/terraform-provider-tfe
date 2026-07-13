# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_azure_oidc_configuration" "example" {
  client_id       = "application-id1"
  subscription_id = "subscription-id1"
  tenant_id       = "tenant-id1"
  organization    = "my-org-name"
}
