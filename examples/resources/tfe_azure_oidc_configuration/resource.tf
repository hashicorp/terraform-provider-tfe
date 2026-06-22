resource "tfe_azure_oidc_configuration" "example" {
  client_id       = "application-id1"
  subscription_id = "subscription-id1"
  tenant_id       = "tenant-id1"
  organization    = "my-org-name"
}
