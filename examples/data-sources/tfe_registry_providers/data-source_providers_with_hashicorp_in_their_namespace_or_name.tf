# Providers with "hashicorp" in their namespace or name

data "tfe_registry_providers" "hashicorp" {
  organization = "my-org-name"
  search       = "hashicorp"
}
