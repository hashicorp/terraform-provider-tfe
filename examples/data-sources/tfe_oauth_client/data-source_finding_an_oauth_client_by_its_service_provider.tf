# Finding an OAuth client by its service provider

data "tfe_oauth_client" "client" {
  organization     = "my-org"
  service_provider = "github"
}
