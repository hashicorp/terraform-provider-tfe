resource "tfe_oauth_client" "test" {
  name             = "my-ado-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://ado.example.com"
  http_url         = "https://ado.example.com"
  oauth_token      = "my-vcs-provider-token"
  private_key      = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  service_provider = "ado_server"
}