resource "tfe_registry_gpg_key" "example" {
  organization = "my-org-name"
  ascii_armor  = file("my-public-key.asc")
}
