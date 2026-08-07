# Write-only configuration

variable "aws_access_key" {
  type      = string
  ephemeral = true
}

variable "aws_secret_key" {
  type      = string
  ephemeral = true
}

resource "tfe_provider_set" "write_only" {
  name            = "example-provider-set-write-only"
  description     = "Reusable provider config with write-only secrets"
  organization    = "example-org"
  provider_source = "registry.terraform.io/hashicorp/aws"
  workspace_ids = [
    "ws-exampleaaaa11111",
    "ws-examplebbbb22222",
  ]

  provider_config_hcl_wo_version = 1
  provider_config_hcl_wo         = <<-EOT
  provider "aws" {
    region     = "us-east-1"
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
  }
  EOT
}
