# Finding an OAuth client by its ID

data "tfe_oauth_client" "client" {
  oauth_client_id = "oc-XXXXXXX"
}
