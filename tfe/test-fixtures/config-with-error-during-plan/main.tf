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
