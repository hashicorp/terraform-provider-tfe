# Azure DevOps Services usage with an organization-scoped personal access token

resource "tfe_oauth_client" "test" {
  name             = "my-ado-services-oauth-client"
  organization     = "my-org-name"
  ado_org_name     = "my-ado-organization"
  api_url          = "https://app.vssps.visualstudio.com"
  http_url         = "https://dev.azure.com"
  oauth_token      = "my-organization-scoped-personal-access-token"
  service_provider = "ado_services"
}
