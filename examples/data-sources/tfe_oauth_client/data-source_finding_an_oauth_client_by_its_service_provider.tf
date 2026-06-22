data "tfe_oauth_client" "client" {
  organization     = "my-org"
  service_provider = "github"
}
