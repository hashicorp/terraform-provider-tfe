# Create a private provider version with SHASUMS files

resource "tfe_registry_provider_version" "example" {
  organization     = tfe_registry_provider.example.organization
  name             = tfe_registry_provider.example.name
  version          = "1.0.0"
  key_id           = tfe_registry_gpg_key.example.id
  protocols        = ["5.0"]
  shasums_file     = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_SHA256SUMS"
  shasums_sig_file = "https://releases.example.com/my-provider/1.0.0/my-provider_1.0.0_SHA256SUMS.sig"
}
