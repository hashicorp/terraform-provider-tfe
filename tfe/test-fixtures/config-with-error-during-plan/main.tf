# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.14.9"
}

variable "name_length" {
  default = 4
  validation = {
    condition     = var.name_length > 10
    error_message = "Name length must be greater than 10"
  }
}
