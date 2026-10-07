# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# Finding an OAuth client by its name

data "tfe_oauth_client" "client" {
  organization = "my-org"
  name         = "my-oauth-client"
}
