# Bitbucket Data Center Usage

resource "tfe_oauth_client" "test" {
  name             = "my-bbdc-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://bbdc.example.com"
  http_url         = "https://bbdc.example.com"
  key              = "<consumer key>"
  secret           = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  rsa_public_key   = "-----BEGIN PUBLIC KEY-----\ncontent\n-----END PUBLIC KEY-----"
  service_provider = "bitbucket_data_center"
}
