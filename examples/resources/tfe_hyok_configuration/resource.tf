# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_gcp_oidc_configuration" "gcp" {
  service_account_email  = "myemail@gmail.com"
  project_number         = "11111111"
  workload_provider_name = "projects/1/locations/global/workloadIdentityPools/1/providers/1"
  organization           = "my-org-name"
}

resource "tfe_hyok_configuration" "example" {
  organization            = "my-hyok-org"
  name                    = "my-key-name"
  kek_id                  = "key1"
  agent_pool_id           = "apool-MFtsuFxHkC9pCRgB"
  oidc_configuration_id   = tfe_gcp_oidc_configuration.gcp.id
  oidc_configuration_type = "gcp"

  kms_options {
    key_location = "global"
    key_ring_id  = "example-key-ring"
  }
}
