# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0



variable "tf_api_token" {
  description = "The API token used should be able to create and configure workspaces variables"
}

variable "tf_hostname" {
  description = "The Terraform Cloud or Enterprise hostname."
  default = "app.terraform.io"
}

provider "tfe" {
  token = var.tf_api_token
  hostname = var.tf_hostname
}

provider "random" {}