# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# This configuration creates and manages workspaces in Terraform Cloud /
# Enterprise. Workspaces and the variables that should be set on them come from
# two maps, and can come from additional sources as well, such as individual
# variable resource blocks. The maps are merged, and then the result is
# iterated over using for_each.

variable "tf_organization" {
  description = "The Terraform Cloud or Enterprise organization under which all operations should be performed."
  type = string
}

variable "vcs_repo_identifier" {
  description = <<-EOT
  The format of VCS repo identifier might differ depending on the VCS provider,
  see https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/workspace
  EOT
  type = string
}

variable "vcs_token" {
  description = "The VCS token should correspond to an API token that can create OAuth clients."
  type = string
}

variable "vars_mapped_by_workspace_name" {
    description = <<-EOT
    This is the map of workspaces and variables. A workspace is created for each
    top level key and then variables are set on the workspace.
    EOT
    type = any
}

variable "additional_vars" {
  description = "This is a map of additional variables intended to be set in specific workspaces."
  type = any
  default = {
    customer_1_workspace = {
      i_am_sensitive_tf_var = {
        value = "i am sensitive"
        sensitive = true
      }
    }
  }
}

variable "default_var_category" {
  description = "Default category for variables being set in managed workspaces unless specified"
  default = "terraform"
  type = string
}

variable "default_var_hcl" {
  description = "By default, variables being set in managed workspaces will not be interpreted as hcl values"
  default = false
  type = bool
}

variable "default_var_sensitive" {
  description = "By default, variables being set in managed workspaces will be non-sensitive"
  default = false
  type = bool
}

locals {
  # Flatten the workspaces variable nested structure for later iteration in the managed_var resource. Adapted from:
  # https://developer.hashicorp.com/terraform/language/functions/flatten#flattening-nested-structures-for-for_each

  # Results in a list of maps, that contains all the information
  # required to manage that workspace variable.
  #   [{
  #     ws            = ws_name
  #     var_key       = name
  #     var_value     = value
  #     var_category  = string
  #     var_hcl       = true/false
  #     var_sensitive = true/false
  #     ws_id         = <tfe_workspace>.id
  #   }...]
  ws_variables = flatten([
    for ws_name, variables in var.vars_mapped_by_workspace_name : [
      for var_name, var_attrs in merge(variables, lookup(var.additional_vars, ws_name, {})) : {
        ws            = ws_name
        var_key       = var_name
        var_value     = var_attrs["value"]
        var_category  = lookup(var_attrs, "category",  var.default_var_category)
        var_hcl       = lookup(var_attrs, "hcl",       var.default_var_hcl)
        var_sensitive = lookup(var_attrs, "sensitive", var.default_var_sensitive)
        ws_id         = tfe_workspace.managed_ws[ws_name].id
      }
    ]
  ])
}

# This example oauth connection assumes the VCS provider is Github.
resource "tfe_oauth_client" "gh" {
  organization     = var.tf_organization
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = var.vcs_token
  service_provider = "github"
}

resource "tfe_workspace" "managed_ws" {
  description = "Create all workspaces specified in the input workspaces map"
  for_each = var.vars_mapped_by_workspace_name

  name = each.key
  organization = var.tf_organization
  auto_apply = true
  force_delete = true
  vcs_repo {
    identifier = var.vcs_repo_identifier
    oauth_token_id = tfe_oauth_client.gh.oauth_token_id
  }
}

resource "tfe_variable" "managed_var" {
  # The for_each expression expects a map or a set of strings,
  # see https://developer.hashicorp.com/terraform/language/meta-arguments/for_each
  # Accordingly, the flattened list of maps, local.ws_variables, is converted to a
  # map with unique key format "workspace_name_variable_name".
  # {
  #   customer_1_workspace_var1 = {
  #     ws            = ws_name
  #     var_key       = name
  #     var_value     = value
  #     var_category  = string
  #     var_hcl       = true/false
  #     var_sensitive = true/false
  #     ws_id         = <tfe_workspace>.id
  #   }
  #   customer_1_workspace_var2 = {
  #     ...
  #   }
  #   ...
  # }

  for_each = {
    for v in local.ws_variables : "${v.ws}.${v.var_key}" => v
  }

  workspace_id = each.value.ws_id
  key          = each.value.var_key
  value        = each.value.var_value
  category     = each.value.var_category
  hcl          = each.value.var_hcl
  sensitive    = each.value.var_sensitive
}

resource "random_pet" "a_dynamic_value" {
  length = 5
}

resource "tfe_variable" "managed_customized_var" {
  description  = "Create dynamic variable for specific workspaces"
  for_each     = tfe_workspace.managed_ws

  workspace_id = each.value.id
  key          = "a_customized_var"
  value        = random_pet.a_dynamic_value.id
  category     = "terraform"
  hcl          = false
  sensitive    = false
}
