## Managing workspaces and variables

This configuration provides an example of how to manage Terraform Cloud / Enterprise workspaces and the variables set on those workspaces using the `tfe` provider.

There are two configurations supplied in this example. The `manager` configuration uses the `tfe` provider to create workspaces and variables. The `managed` configuration is the configuration that will be associated with the workspaces being created. This `managed` workspace represents what would do the actual deployment of the resources of interest, which could represent environments (dev/test/prod), customers, and so on. The use case is for any configuration that needs to be deployed where the only difference is the variables supplied.

### Initial setup and information gathering

Successful configuration of this example requires several things to be prepared in advance.

* Create or join a Terraform Cloud or Enterprise [organization](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/organizations#creating-organizations).
* Create a Terraform Cloud or Enterprise token to use with the `tfe` [provider](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs). While other token types can be used, a [User Token](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/users#api-tokens) is recommended in the beginning so that the provider has the same access as the user experimenting with this configuration.
* Create a [Github token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) that has access to create and manage OAuth clients. A personal access token with full repo scope should be sufficient. This example uses Github, so the configuration will need to be modified to work with other VCS providers.

Have the following information handy for the next steps:

* TFC or TFE organization name
* TFC or TFE token
* Github organization name
* Github repository names (`managed` and `manager`, but you could, and eventually probably will, rename them)
* Github token

### Setting up initial bootstrap infrastructure

* Create two Github repositories named `manager` and `managed`.
* Add the `manager` configuration to the `manager` repository.
* Add the `managed` configuration to the `managed` repository.
* Create a TFC/TFE workspace named `manager` in the organization that was created or joined.
  - If there is no existing connection for Github configured, select to [add a connection](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/vcs).
  - With Github connected, choose the `manager` repository from the list.
  - Use the default options for the remaining settings to complete the workspace setup.

### Manager workspace variables

The `manager` workspace requires certain Terraform variables to be set on the [Variables page in the UI](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables#managing-variables-in-the-ui). All variables should be "Terraform Variables". There are no "Environment Variables" required.

* `tf_hostname` - The Terraform Cloud or Enterprise hostname. Defaults to app.terraform.io
* `tf_api_token` - The TFC/TFE user token that was generated. Mark this as sensitive.
* `tf_organization` - The TFC/TFE organization name where the managed workspaces should exist. This will most commonly be the organization that was created or joined, where the `manager` workspace also exists. The token supplied to perform the operations should have sufficient access to the organization.
* `vcs_repo_identifier` - A reference to the `managed` repository in the format `<github-organization>/<repository>`. The format of the VCS repo identifier might differ depending on the VCS provider, see [tfe_workspace](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/workspace)
* `vcs_token` - The Github personal access token that can create OAuth connections. Mark this as sensitive.

### Run to create managed workspaces

With the variables populated, perform a run in the `manager` workspace. The plan should indicate that it will create two managed workspaces: `customer_1_workspace` and `customer_2_workspace`. Additionally, an OAuth connection for these repositories to Github will be created, and the specified variables will be set on both workspaces.

### Further details

As mentioned, the `manager` workspace uses a map to obtain workspace names and the variables that should be set on each workspace. This variable, `vars_mapped_by_workspace_name`, is loaded from `workspaces.auto.tfvars` and contains all of the variables whose values can be stored safely in the `manager` repository.

There is another map that is used to supply more sensitive data, `additional_vars`. Variables specified in `vars_mapped_by_workspace_name` and `additional_vars` will be merged during the run. The value of `additional_vars` will need to be stored elsewhere for reference when it needs to be changed.

Default values for the variables are set in the `managed` configuration, so not every variable needs to be defined in the map for every workspace.

The `vars_mapped_by_workspace_name` map demonstrates only one of many ways in which variables can be set on workspaces. The map approach provides a concise way to set many variables. The `tfe_variable.managed_customized_var` resource in the `manager` configuration demonstrates how to add another variable that is not a part of the `workspaces` map. Other approaches and sources for variable names and values can be used so long as there are no variable name collisions.
