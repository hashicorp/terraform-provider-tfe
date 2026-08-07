# Finding an OAuth client by its name

data "tfe_oauth_client" "client" {
  organization = "my-org"
  name         = "my-oauth-client"
}
