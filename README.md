# Terraform Enterprise Provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-tfe`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-tfe
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-tfe
$ make build
```

To use this provider binary, you have a few different options:
* You can copy the provider binary to your `~/.terraform.d/plugins` directory by running the following command:
   ```sh
   $ mv $GOPATH/bin/terraform-provider-tfe ~/.terraform.d/plugins
   ```
* You can create your test Terraform configurations in the same directory as your provider binary or you can copy the provider binary into the same directory as your test configurations.
* You can copy the provider binary into the same locations as your `terraform` binary.

To learn more about using a local build of a provider, you can look at the [documentation on writing custom providers](https://www.terraform.io/docs/extend/writing-custom-providers.html#invoking-the-provider) and the [documentation on how Terraform plugin discovery works](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery)

## Using the provider

For production use, you should constrain the acceptable provider versions via configuration,
to ensure that new versions with breaking changes will not be automatically installed by 
`terraform init` in future. As this provider is still at version zero, you should constrain 
the acceptable provider versions on the minor version.

If you are using Terraform CLI version 0.12.x, you can constrain this provider to 0.15.x versions 
by adding a `required_providers` block inside a `terraform` block.
```
terraform {
  required_providers {
    tfe = "~> 0.15.0"
  }
}
```

If you are using Terraform CLI version 0.11.x, you can constrain this provider to 0.15.x versions 
by adding the version constraint to the tfe provider block.
```
provider "tfe" {
  version = "~> 0.15.0"
  ...
}
```

For more information on constraining provider versions, see the 
[provider versions documentation](https://www.terraform.io/docs/configuration/providers.html#provider-versions).

If you're building the provider, follow the instructions to
[install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin)
After placing it into your plugins directory,  run `terraform init` to initialize it.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed
on your machine (version 1.11+ is *required*). You'll also need to correctly setup a
[GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary
in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-tfe
...
```

## Referencing a local version of `go-tfe`

You may want to create configs or run tests against a local version of `go-tfe`. Add the following line to `go.mod` above the require statement, using your local path to `go-tfe`.

```
replace github.com/hashicorp/go-tfe => /path-to-local-repo/go-tfe
```

## Testing

### 1. (Optional) Create repositories for policy sets, registry modules, and workspaces

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

## Updating the Changelog

Only update the `Unreleased` section. Make sure you change the unreleased tag to an appropriate version, using [Semantic Versioning](https://semver.org/) as a guideline.

Please use the template below when updating the changelog:
```
<change category>:
* **New Resource:** `name_of_new_resource` ([#123](link-to-PR))
* r/tfe_resource: description of change or bug fix ([#124](link-to-PR))
```

### Change categories

- BREAKING CHANGES: Use this for any changes that aren't backwards compatible. Include details on how to handle these changes.
- FEATURES: Use this for any large new features added.
- ENHANCEMENTS: Use this for smaller new features added.
- BUG FIXES: Use this for any bugs that were fixed.
- NOTES: Use this section if you need to include any additional notes on things like upgrading, upcoming deprecations, or any other information you might want to highlight.
