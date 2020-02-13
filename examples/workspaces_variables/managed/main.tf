# Defaults are given in this configuration and also using the some.auto.tfvars.
# Both will be overridden by what is set in the TF Cloud / Enterprise Variables
# page using the manager configuration.

# Edit and uncomment this backend configuration to enable connecting to the
# different managed workspaces created by the manager from your workstation.
# Connecting to the backend from your workstation is not required in order to
# use this configuration from within TF Cloud / Enterprise.

# Note that the prefix should end in a literal hyphen. For the repo 'managed',
# the prefix should be 'managed-'

/*
terraform {
  backend "remote" {
    organization = "YOUR_ORGANIZATION"
    workspaces = {
      prefix = "managed-"
    }
  }
}
*/

#
# Variables
#

variable "sens_tf_var" {
  default = ""
}

variable "sens_env_var" {
  default = ""
}

variable "a_string" {
  type = string
  default = "a default string"
}

# Value is initially supplied via some.auto.tfvars
variable "a_number" {
  type = number
}

variable "a_list" {
  type = list
  default = []
}

# Value is initially supplied via some.auto.tfvars
variable "a_map" {
  type = map
}

variable "a_single_var" {
  default = ""
}

#
# Output all of the variables for demonstration purposes
#

output "sens_tf_var" {
  value = var.sens_tf_var
}

output "a_string" {
  value = var.a_string
}

output "a_number" {
  value = var.a_number
}

output "a_list" {
  value = var.a_list
}

output "a_map" {
  value = var.a_map
}

output "a_single_var" {
  value = var.a_single_var
}

# Print the value of the environment variables 
resource "null_resource" "env_vars" {
  provisioner "local-exec" {
    command = <<EOT
      echo an_env = $an_env
      echo another_env = $another_env
      echo sens_env_var = $sens_env_var
    EOT
  }
}
