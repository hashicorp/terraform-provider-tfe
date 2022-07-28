resource "tfe_oauth_client" "test" {
  name             = "my-bbs-oauth-client"
  organization     = "my-org-name"
  api_url          = "https://bbs.example.com"
  http_url         = "https://bss.example.com"
  key              = "<consumer key>"
  secret           = "-----BEGIN RSA PRIVATE KEY-----\ncontent\n-----END RSA PRIVATE KEY-----"
  rsa_public_key   = "-----BEGIN PUBLIC KEY-----\ncontent\n-----END PUBLIC KEY-----"
  service_provider = "bitbucket_server"
}