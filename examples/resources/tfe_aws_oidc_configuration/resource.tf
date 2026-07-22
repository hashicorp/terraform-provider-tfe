# Basic usage

resource "tfe_aws_oidc_configuration" "example" {
  role_arn     = "arn:aws:iam::111111111111:role/example-role-arn"
  organization = "my-org-name"
}
