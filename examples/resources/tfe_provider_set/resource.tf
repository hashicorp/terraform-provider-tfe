# Basic usage

resource "tfe_provider_set" "standard" {
  name            = "example-provider-set"
  description     = "Reusable provider config for selected workspaces"
  organization    = "example-org"
  provider_source = "registry.terraform.io/hashicorp/aws"
  priority        = true
  workspace_ids = [
    "ws-exampleaaaa11111",
    "ws-examplebbbb22222",
  ]

  provider_config_hcl = <<-EOT
  provider "aws" {
    region = "us-east-1"
  }
  EOT
}
