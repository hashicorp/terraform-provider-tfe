# Azure DevOps Server Usage
# Note that this resource requires a private key when creating Azure DevOps Server OAuth clients.
# Documentation for HCP Terraform and Terraform Enterprise setup can be found here: https://developer.hashicorp.com/terraform/cloud-docs/vcs/azure-devops-server

resource "tfe_oauth_client" "test" {
  name             = "my-ado-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://ado.example.com"
  http_url         = "https://ado.example.com"
  oauth_token      = "my-vcs-provider-token"
  private_key      = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  service_provider = "ado_server"
}
