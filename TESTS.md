# Running tests

Running all the tests for this provider requires access to Terraform Cloud with
a full feature set; most tests can be run against your own installation of
Terraform Enterprise.

## 1. (Optional) Create repositories for policy sets, registry modules, and workspaces

If you are planning to run the full suite of tests or work on policy sets, registry modules, or workspaces, you'll need to set up repositories for them in GitHub.

Your policy set repository will need the following: 
1. A policy set stored in a subdirectory
1. A branch other than master

Your registry module repository will need to be a [valid module](https://www.terraform.io/docs/cloud/registry/publish.html#preparing-a-module-repository).
It will need the following: 
1. To be named `terraform-<PROVIDER>-<NAME>`
1. At least one valid SemVer tag in the format `x.y.z`
[terraform-random-module](ttps://github.com/caseylang/terraform-random-module) is a good example repo.

Your workspace repository will need the following: 
1. A branch other than master
   
### 2. Set up environment variables

To run all tests, you will need to set the following environment variables:

##### Required:
A hostname and token must be provided in order to run the acceptance tests. By
default, these are loaded from the the `credentials` in the [CLI config
file](https://www.terraform.io/docs/commands/cli-config.html). You can override
these values with the environment variables specified below: 

1. `TFE_HOSTNAME` - URL of a Terraform Cloud or Terraform Enterprise instance to be used for testing, without the scheme. Example: `tfe.local`
1. `TFE_TOKEN` - A [user API token](https://www.terraform.io/docs/cloud/users-teams-organizations/users.html#api-tokens) for an administrator account on the Terraform Cloud or Terraform Enterprise instance being used for testing.

##### Optional:
1. `TFE_USER1` and `TFE_USER2`: The usernames of two pre-existing users on the Terraform Cloud or Terraform Enterprise instance being used for testing. Required for running team membership tests.
1. `GITHUB_TOKEN` - [GitHub personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line). Used to establish a VCS provider connection.
1. `GITHUB_POLICY_SET_IDENTIFIER` - GitHub policy set repository identifier in the format `username/repository`. Required for running policy set tests.
1. `GITHUB_POLICY_SET_BRANCH`: A GitHub branch for the repository specified by `GITHUB_POLICY_SET_IDENTIFIER`. Required for running policy set tests.
1. `GITHUB_POLICY_SET_PATH`: A GitHub subdirectory for the repository specified by `GITHUB_POLICY_SET_IDENTIFIER`. Required for running policy set tests.
1. `GITHUB_REGISTRY_MODULE_IDENTIFIER` - GitHub registry module repository identifier in the format `username/repository`. Required for running registry module tests.
1. `GITHUB_WORKSPACE_IDENTIFIER` - GitHub workspace repository identifier in the format `username/repository`. Required for running workspace tests.
1. `GITHUB_WORKSPACE_BRANCH`: A GitHub branch for the repository specified by `GITHUB_WORKSPACE_IDENTIFIER`. Required for running workspace tests.

You can set your environment variables up however you prefer. The following are instructions for setting up environment variables using [envchain](https://github.com/sorah/envchain).
   1. Make sure you have envchain installed. [Instructions for this can be found in the envchain README](https://github.com/sorah/envchain#installation).
   1. Pick a namespace for storing your environment variables. I suggest `terraform-provider-tfe` or something similar.
   1. For each environment variable you need to set, run the following command:
      ```sh
      envchain --set YOUR_NAMESPACE_HERE ENVIRONMENT_VARIABLE_HERE
      ```
      **OR**
    
      Set all of the environment variables at once with the following command:
      ```sh
      envchain --set YOUR_NAMESPACE_HERE TFE_HOSTNAME TFE_TOKEN TFE_USER1 TFE_USER2 GITHUB_TOKEN GITHUB_POLICY_SET_IDENTIFIER GITHUB_POLICY_SET_BRANCH GITHUB_POLICY_SET_PATH GITHUB_REGISTRY_MODULE_IDENTIFIER GITHUB_WORKSPACE_IDENTIFIER GITHUB_WORKSPACE_BRANCH
      ```
  
### 3. Run the tests

#### Running the provider tests

##### With envchain:
```sh
$ envchain YOUR_NAMESPACE_HERE make test
```

##### Without envchain:
```sh
$ make test
```

#### Running all the acceptance tests

##### With envchain:
```sh
$ envchain YOUR_NAMESPACE_HERE make testacc
```

##### Without envchain:
```sh
$ make testacc
```

#### Running specific acceptance tests 

The commands below use notification configurations as an example.

##### With envchain:
```sh
$ TESTARGS="-run TestAccTFENotificationConfiguration" envchain YOUR_NAMESPACE_HERE make testacc
```

##### Without envchain:
```sh
$ TESTARGS="-run TestAccTFENotificationConfiguration" make testacc
```   

