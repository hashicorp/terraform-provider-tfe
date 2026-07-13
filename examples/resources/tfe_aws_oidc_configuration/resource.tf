# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_aws_oidc_configuration" "example" {
  role_arn     = "arn:aws:iam::111111111111:role/example-role-arn"
  organization = "my-org-name"
}
