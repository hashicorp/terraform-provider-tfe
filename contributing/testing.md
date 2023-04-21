# Running tests

Running all the tests for this provider requires access to Terraform Cloud with
a full feature set; most tests can be run against your own installation of
Terraform Enterprise.

## 1. (Optional) Create repositories for policy sets, registry modules, and workspaces

If you are planning to run the full suite of tests or work on policy sets, registry modules, or workspaces, you'll need to set up repositories for them in GitHub.

Your policy set repository will need the following:
1. A policy set stored in a subdirectory
1. A branch other than `main`.

Your registry module repository will need to be a [valid module](https://developer.hashicorp.com/terraform/cloud-docs/registry/publish-modules#preparing-a-module-repository).
It will need the following:
1. To be named `terraform-<PROVIDER>-<NAME>`
1. At least one valid SemVer tag in the format `x.y.z`
[terraform-random-module](https://github.com/caseylang/terraform-random-module) is a good example repo.

Your workspace repository will need the following:
1. A branch other than `main`.

### 2. Set up environment variables

To run all tests, you will need to set the following environment variables:

##### Required:
A hostname and token must be provided in order to run the acceptance tests. By
default, these are loaded from the `credentials` in the [CLI config
file](https://developer.hashicorp.com/terraform/cli/config/config-file). You can override
these values with the environment variables specified below:

1. `TFE_HOSTNAME` - URL of a Terraform Cloud or Terraform Enterprise instance to be used for testing, without the scheme. Example: `tfe.local`
1. `TFE_TOKEN` - A [user API token](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/users#tokens) for an administrator account on the Terraform Cloud or Terraform Enterprise instance being used for testing.

##### Optional:
1. `TFE_USER1` and `TFE_USER2`: The usernames of two pre-existing users on the Terraform Cloud or Terraform Enterprise instance being used for testing. Required for running team membership tests.
2. `GITHUB_TOKEN` - [GitHub personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line). Used to establish a VCS provider connection.
3. `GITHUB_POLICY_SET_IDENTIFIER` - GitHub policy set repository identifier in the format `username/repository`. Required for running policy set tests.
4. `GITHUB_POLICY_SET_BRANCH`: A GitHub branch for the repository specified by `GITHUB_POLICY_SET_IDENTIFIER`. Required for running policy set tests.
5. `GITHUB_POLICY_SET_PATH`: A GitHub subdirectory for the repository specified by `GITHUB_POLICY_SET_IDENTIFIER`. Required for running policy set tests.
6. `GITHUB_REGISTRY_MODULE_IDENTIFIER` - GitHub registry module repository identifier in the format `username/repository`. Required for running registry module tests.
7. `GITHUB_WORKSPACE_IDENTIFIER` - GitHub workspace repository identifier in the format `username/repository`. Required for running workspace tests.
8. `GITHUB_WORKSPACE_BRANCH`: A GitHub branch for the repository specified by `GITHUB_WORKSPACE_IDENTIFIER`. Required for running workspace tests.
9. `ENABLE_TFE` - Some tests cover features available only in Terraform Cloud. To skip these tests when running against a Terraform Enterprise instance, set `ENABLE_TFE=1`.
10. `RUN_TASKS_URL` - External URL to use for testing Run Tasks operations, for example `RUN_TASKS_URL=http://somewhere.local:8080/pass`. Required for running run tasks tests.
11. `GITHUB_APP_INSTALLATION_ID` - GitHub App installation internal id in the format `ghain-xxxxxxx`. Required for running any tests that use GitHub App VCS (workspace, policy sets, registry module).
12. `GITHUB_APP_INSTALLATION_NAME` - GitHub App installation name. Required for running tfe_github_app_installation data source test.

**Note:** In order to run integration tests for **Paid** features you will need a token `TFE_TOKEN` with TFC/E administrator privileges, otherwise the attempt to upgrade an organization's feature set will fail.

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
      envchain --set YOUR_NAMESPACE_HERE TFE_HOSTNAME TFE_TOKEN TFE_USER1 TFE_USER2 GITHUB_TOKEN GITHUB_POLICY_SET_IDENTIFIER GITHUB_POLICY_SET_BRANCH GITHUB_POLICY_SET_PATH GITHUB_REGISTRY_MODULE_IDENTIFIER GITHUB_WORKSPACE_IDENTIFIER GITHUB_WORKSPACE_BRANCH GITHUB_APP_INSTALLATION_ID GITHUB_APP_INSTALLATION_NAME
      ```

### 3. Run the tests

There are two types of tests one can run for the provider: acceptance tests and unit tests. You can run acceptance tests using the Makefile target `testacc` and unit tests using the Makefile target `test`. Typically, when you write a test for a particular resource or data source it will be referred to as an acceptance test. On the other hand, unit tests are reserved for resource helpers or provider specific logic. These are semantics used by the Terraform Plugin SDKv2 and are maintained here for consistency, learn more about [Acceptance Tests](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests). Furthermore, resource tests make use of the Terraform Plugin SDKv2 helper, [resource.Test()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#Test), which requires the environment variable `TF_ACC` to be set in order to run.

**Note**: The difference between `make testacc` and `make test` is whether `TF_ACC=1` is set or not. However, you can still run unit tests using the `testacc` target.

#### Run all acceptance tests

##### With envchain:
```sh
$ envchain YOUR_NAMESPACE_HERE make testacc
```

##### Without envchain:
```sh
$ make testacc
```

#### Run a specific acceptance test

The commands below use notification configurations as an example.

##### With envchain:
```sh
$ TESTARGS="-run TestAccTFENotificationConfiguration" envchain YOUR_NAMESPACE_HERE make testacc
```

##### Without envchain:
```sh
$ TESTARGS="-run TestAccTFENotificationConfiguration" make testacc
```

#### Run all unit tests

##### With envchain:
```sh
$ envchain YOUR_NAMESPACE_HERE make test
```

##### Without envchain:
```sh
$ make test
```

#### Run a specific unit test

The commands below test the organization run task helper as an example.

##### With envchain:
```sh
$ TESTARGS="-run TestFetchOrganizationRunTask" envchain YOUR_NAMESPACE_HERE make test
```

##### Without envchain:
```sh
$ TESTARGS="-run TestFetchOrganizationRunTask" make test
```

