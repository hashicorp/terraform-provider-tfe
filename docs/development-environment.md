# Development Environment Setup

## Prerequisites

- [Go 1.17+](https://golang.org/doc/install) (to build the provider and run the tests)
- [Terraform 0.14+](https://developer.hashicorp.com/terraform/downloads) (to run the tests)
- [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) (to run code checks locally)

## Building the Provider

Clone the repository, enter the directory, and build the provider:

```sh
$ git clone git@github.com:hashicorp/terraform-provider-tfe
$ cd terraform-provider-tfe
$ make
```

This will build the provider and put the binary in the project directory. To use the compiled binary, you have several different options (this list is not exhaustive):

##### Using CLI config to provide a dev override (Using Terraform v0.14+)

Use the rule `make devoverride` to generate a CLI config containing a dev override provider installation. This command will print a variable export that can be copied and pasted into a shell session while testing with terraform. To automatically export this override, use `eval $(make devoverride)` This command will override any custom Terraform CLI config file path you have previously defined.

Example usage:

```sh
$ eval $(make devoverride)
$ cd ../example-terraform-config
$ terraform init
```

##### Using Terraform 0.13+

You can use a filesystem mirror (either one of the [implied local mirror directories](https://developer.hashicorp.com/terraform/cli/config/config-file#implied-local-mirror-directories) for your platform
or by [configuring your own](https://developer.hashicorp.com/terraform/cli/config/config-file#explicit-installation-method-configuration)).

See the [Provider Requirements](https://developer.hashicorp.com/terraform/language/providers/requirements) documentation for more information.

##### Using Terraform 0.12

* You can copy the provider binary to your `~/.terraform.d/plugins` directory.
* You can create your test Terraform configurations in the same directory as your provider binary or you can copy the provider binary into the same directory as your test configurations.
* You can copy the provider binary into the same location as your `terraform` binary.

## Running the Tests

The provider is mainly tested using a suite of acceptance tests that run against an internal instance of Terraform Cloud. We also test against Terraform Enterprise prior to release.

To run the acceptance tests, run `make testacc`

```sh
$ make testacc
```

### Referencing a local version of `go-tfe`

You may want to create configs or run tests against a local version of `go-tfe`. The following command can be used to temporarily override the go-tfe module to your local version:

```
go mod edit -replace github.com/hashicorp/go-tfe=../go-tfe
```

### Running the Code Checks Locally

This repository uses golangci-lint to check for common style issues. To run them before submitting your PR, run `make lint`

```sh
make lint
```

Optionally, to integrate golangci-lint into your editor, see [golangci-lint editor integration](https://golangci-lint.run/usage/integrations/)
