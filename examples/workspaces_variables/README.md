## Managing workspaces and variables

This configuration provides one example of how to manage Terraform Cloud / Enterprise workspaces and the variables set on those workspaces using the `tfe` provider.

There are two configurations supplied in this example. The `manager` configuration uses the `tfe` provider to crate workspaces and variables. The `managed` configuration is the configuration that will be associated with the workspaces being created. This `managed` workspace represents what would do the actual deployment of the resources of interest, and would represent environments (dev/test/prod), customers, and so on. Any configuration that needs to be deployed where the only difference is the variables that are supplied.

### Initial setup and information gathering

Successful configuration of this example requires several things to be prepared in advance.

* Create or join a Terraform Cloud or Enterprise [organization](https://www.terraform.io/docs/cloud/users-teams-organizations/organizations.html#creating-organizations).
* Create a Terraform Cloud or Enterprise token to use with the `tfe` [provider](https://www.terraform.io/docs/providers/tfe/index.html). While other token types can be used, a [User Token](https://www.terraform.io/docs/cloud/users-teams-organizations/users.html#api-tokens) is recommended in the beginning so that the provider has the same access as the user experimenting with this configuration.
* Create a [Github token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) that has access to create and manage OAuth clients. A personal access token should be sufficient. This example uses Github, so the configuration will need to be modified to work with other VCS providers.

Have the following information handy for the next steps:

* TFC or TFE organization name
* TFC or TFE token
* Github organization name
* Github repository names (`managed` and `manager`, but you could, and eventually probably will, rename them)
* Github token

### Setting up initial bootstrap infrastructure

* Create two Github repositories named `manager` and `managed`.
* Add the manager configuration to the manager repository.
* Add the managed configuration to the managed repository.
* Create a TFC/TFE workspace named `manager` in the organization that was created or joined.
  - If there is no existing connection for Github configured, select to [add a connection](https://www.terraform.io/docs/cloud/workspaces/vcs.html).
  - With Github connected, choose the manager repository from the list.
  - Use the default options for the remaining settings to complete the workspace setup.

### Manager workspace variables

The manager workspace requires certain Terraform variables to be set on the [Variables page in the UI](https://www.terraform.io/docs/cloud/workspaces/variables.html#managing-variables-in-the-ui). All variables should be "Terraform Variables". There are no "Environment Variables" required.

* `tf_api_token` - The TFC/TFE user token that was generated. Mark this sensitive.
* `org` - The TFC/TFE organization name where the managed workspaces should exist. This will most commonly be the organization that was created or joined, where the `manager` workspace also exists, but that is not required. The only requirement is that the token supplied to perform the operations has sufficient access to the organization.
* `vcs_org` - The Github organization that the `manager` and `managed` repositories are under.
* `vcs_repo` - This should be `managed`. This is a variable because eventually you will likely want to rename it to something specific to your use, such as `customer` or `web_application`.
* `vcs_token` - The Github personal access token that can create OAuth connections. Mark this sensitive.

### Run to create managed workspaces

With the variables populated, perform a run in the `manager` workspace. The plan should indicate that it will create two workspaces: `managed-ws1` and `managed-ws2`. Additionally, an OAuth connection for these repositories to Github will be created, and a number of variables will be set on both workspaces.

### Further details

As mentioned, the `manager` workspace uses a map to obtain workspace names and the variables that should be set on each workspace. This variable, `workspaces`, is loaded from `workspaces.auto.tfvars` and contains all of the variables whose values can be stored safely in the `manager` repository.

There is another map that can be used to supply more sensitive data, `addtl_vars`. To use `addtl_vars`, create a Terraform Variable in the `manager` workspace called `addtl_vars`, enabling the HCL and Sensitive options. The `addtl_vars` variable should have the same structure as the `workspaces` variable, and follow the example supplied. The two maps will be merged during the run. The value of `addtl_vars` will need to be stored elsewhere for reference when it needs to be changed.

Default values for the variables are set in the `managed` configuration, so not every variable needs to be defined in the map for every workspace.

The `workspaces` map demonstrates only one of many ways in which variables can be set on workspaces. The map approach provides a concise way to set many variables. The `tfe_variable.managed_single_var` resource in the `manager` configuration demonstrates how to add another variable that is not a part of the `workspaces` map. Other approaches and sources for variable names and values can be used so long as there are no variable name collisions.
