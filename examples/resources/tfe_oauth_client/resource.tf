resource "tfe_oauth_client" "test" {
  name                = "my-github-oauth-client"
  organization        = "my-org-name"
  api_url             = "https://api.github.com"
  http_url            = "https://github.com"
  oauth_token         = "my-vcs-provider-token"
  service_provider    = "github"
  organization_scoped = true
}
