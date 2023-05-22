# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
  required_version = ">= 0.14.9"
}

resource "random_pet" "always_new" {
  keepers = {
    uuid = uuid()
  }
  length = 5
}

output "pet" {
  value = { name_of_pet : random_pet.always_new.id }
}