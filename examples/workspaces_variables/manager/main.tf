# This configuration creates and manages workspaces in Terraform Cloud /
# Enterprise. Workspaces and the variables that should be set on them come from
# two maps, and can come from additional sources as well, such as individual
# variable resource blocks. The maps are merged, and then the result is
# iterated over using for_each.

# Edit and uncomment this block if you wish to connect to the manager workspace
# from your workstation using the remote backend. This is not necessary in
# order to use and run this configuration in Terraform Cloud / Enterprise.
# It is also necessary to create a terraformrc configuration file:
# https://www.terraform.io/docs/commands/cli-config.html
/*
terraform {
  backend "remote" {
    organization = "YOUR_ORG"
    workspaces {
      name = "manager"
    }
  }
}
*/

# By default, variables being set will be Terraform variables
variable "default_var_category" {
  default = "terraform"
}

# By default, variables being set will not be interpreted as hcl values
variable "default_var_hcl" {
  default = false
}

# By default, variables being set will not be sensitive
variable "default_var_sensitive" {
  default = false
}

# The API token used should be able to create and configure workspaces, and
# create and configure workspace variables.
variable "tf_api_token" {}

# The Terraform Cloud or Enterprise organization under which all operations
# should be performed.
variable "org" {}

# The org and repo should correspond to, e.g., github.com/org/repo
variable "vcs_org" {}
variable "vcs_repo" {}

# The vcs token should correspond to an API token that can create OAuth
# clients.
variable "vcs_token" {}

# This is the map of workspaces and variables. A workspace is created for each
# top level key and then variables are set on the workspace.
variable "workspaces" {}

# This is a map of additional variables intended to be set in the Terraform
# Enterprise workspace so that it can be set as sensitive so the values are
# hidden.
variable "addtl_vars" {
  default = {}
}

provider "tfe" {
  token = var.tf_api_token
}

locals {
  # Flatten a nested structure for later iteration in a resource. Adapted from:
  # https://www.terraform.io/docs/configuration/functions/flatten.html#flattening-nested-structures-for-for_each

  # Results in a list that can be used to create a map, where each key
  # represents a workspace variable that needs to be set, and each value
  # contains all of the information required to manage that workspace variable.
  ws_variables = flatten([
    for ws_name, variables in var.workspaces : [
      for var_name, var_attrs in merge(variables, lookup(var.addtl_vars, ws_name, {})) : {
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
  organization     = var.org
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = var.vcs_token
  service_provider = "github"
}

# Create a workspace for each top level key. The workspace will be named after
# the VCS repo and the top level key: "repo-name"
resource "tfe_workspace" "managed_ws" {
  for_each = var.workspaces

  name = "${var.vcs_repo}-${each.key}"
  organization = var.org

  vcs_repo {
    identifier = "${var.vcs_org}/${var.vcs_repo}"
    oauth_token_id = tfe_oauth_client.gh.oauth_token_id
  }
}

resource "tfe_variable" "managed_var" {
  # The for_each expression expects a map. The flattened list of maps,
  # local.ws_variables, contains all of the required information to create all
  # of the variables. What's required, then, is a map where each key/value is a
  # variable to create. The keys need to be unique, so the key used here is
  # "workspace_name.variable_name".  The each.key expression is not used. The
  # each.value map has all of the variable information.

  # So the transformation is from the merged var.workspaces and addtl_vars
  # structure:
  # {
  #   ws1 {
  #     var1 = {
  #       var_key   = name
  #       var_value = value
  #       var_hcl   = true/false
  #       ws_id     = <tfe_workspace>.id
  #     }
  #     var2 = {
  #       ...
  #     }
  #   }
  #   ws2 {
  #     ...
  #   }
  #   ...
  # }
  #
  # To a map of unique keys pointing at each list value in local.ws_variables:
  # {
  #   ws1_var1 = {
  #     ws        = ws_name
  #     var_key   = name
  #     var_value = value
  #     var_hcl   = true/false
  #     ws_id     = <tfe_workspace>.id
  #   }
  #   ws1_var2 = {
  #     ...
  #   }
  #   ws2_var1 = {
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

resource "tfe_variable" "managed_single_var" {
  for_each     = tfe_workspace.managed_ws

  workspace_id = each.value.id
  key          = "a_single_var"
  value        = "not from the workspaces map"
  category     = "terraform"
  hcl          = false
  sensitive    = false
}
