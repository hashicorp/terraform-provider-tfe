# Basic usage

data "tfe_org_max_token_ttl_policy" "example" {
  organization = "my-org-name"
}

output "org_token_ttl_ms" {
  value = data.tfe_org_max_token_ttl_policy.example.org_token_max_ttl_ms
}
