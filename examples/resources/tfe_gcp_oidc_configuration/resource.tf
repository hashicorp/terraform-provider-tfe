resource "tfe_gcp_oidc_configuration" "example" {
  service_account_email  = "myemail@gmail.com"
  project_number         = "11111111"
  workload_provider_name = "projects/1/locations/global/workloadIdentityPools/1/providers/1"
  organization           = "my-org-name"
}
