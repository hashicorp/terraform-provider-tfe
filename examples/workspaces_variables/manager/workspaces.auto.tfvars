# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# If a variable is omitted it will not be set on the variables page. Values can
# be supplied in other ways, such as *.auto.tfvars files, via the variable's
# default value, var.addtl_vars, or manually managed variables.

# Environment variables that are not set default to being unset / empty, as
# that's just a consequence of them being environment variables.

# A variable must have a value, and beyond that can have any of three other
# attributes.
# a_variable = {
#    value = "val1"
#    category = "terraform" (default) or "env"
#    hcl = true or false (default)
#    sensitive = true or false (default)
# }

# Note that lists and maps need to be enclosed in heredoc syntax so that they
# are actually strings, which is what the tfe provider requires.

vars_mapped_by_workspace_name = {
  customer_1_workspace = {
    a_string = {
      value = "val1"
    }

    a_number = {
      value = 3.14
    }

    a_list = {
      value = <<-EOT
        [
          "one",
          "two",
        ]
      EOT
      hcl = true
    }

    a_map = {
      value = <<-EOT
        {
          foo = "bar"
          baz = "qux"
        }
      EOT
      hcl = true
    }

    an_env = {
      value = "an env var"
      category = "env"
    }
  }

  customer_2_workspace = {
    a_number = {
      value = 6.28
    }

    a_map = {
      value = <<-EOT
        {
          foo = "bar"
          baz = "qux"
        }
      EOT
      hcl = true
    }

    another_env = {
      value = "another env var"
      category = "env"
    }
  }
}

