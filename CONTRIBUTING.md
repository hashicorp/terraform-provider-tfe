# Contributing

Thanks for your interest in contributing; we appreciate your help! If you're unsure or afraid of anything, you can
submit a work in progress (WIP) pull request, or file an issue with the parts you know. We'll do our best to guide you
in the right direction, and let you know if there are guidelines we will need to follow. We want people to be able to
participate without fear of doing the wrong thing.

👉 _See [Manually building the provider](#manually-building-the-provider) below._

Other helpful resources:

* [Extending Terraform documentation](https://www.terraform.io/docs/extend/index.html)
* [Terraform Cloud API documentation](https://www.terraform.io/docs/cloud/api/index.html)
* [Package documentation for the Terraform Cloud/Enterprise Go client (go-tfe)](https://pkg.go.dev/github.com/hashicorp/go-tfe)

### Manually building the provider

You might prefer to manually build the provider yourself - perhaps access to the Terraform Registry or the official
release binaries on [releases.hashicorp.com](https://releases.hashicorp.com/terraform-provider-tfe/) are not available
in your operating environment, or you're looking to contribute to the provider and are testing out a custom build.

Building the provider requires [Go](https://golang.org/doc/install) >= 1.16

Clone the repository, enter the directory, and build the provider:

```sh
$ git clone git@github.com:hashicorp/terraform-provider-tfe
$ cd terraform-provider-tfe
$ make build
```

This will build the provider and put the binary in the `$GOPATH/bin` directory. To use the compiled binary, you have several different options (this list is not exhaustive):

##### Using Terraform 0.13+

You can use a filesystem mirror (either one of the [implied local mirror directories](https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories) for your platform
or by [configuring your own](https://www.terraform.io/docs/commands/cli-config.html#explicit-installation-method-configuration)).

See the [Provider Requirements](https://www.terraform.io/docs/configuration/provider-requirements.html) documentation for more information.

##### Using Terraform 0.12

* You can copy the provider binary to your `~/.terraform.d/plugins` directory.
* You can create your test Terraform configurations in the same directory as your provider binary or you can copy the provider binary into the same directory as your test configurations.
* You can copy the provider binary into the same location as your `terraform` binary.

## Referencing a local version of `go-tfe`

You may want to create configs or run tests against a local version of `go-tfe`. Add the following line to `go.mod` above the require statement, using your local path to `go-tfe`:

```
replace github.com/hashicorp/go-tfe => /path-to-local-repo/go-tfe
```

### Running the Linters Locally

1. Ensure you have [installed golangci-lint](https://golangci-lint.run/usage/install/#local-installation)
2. From the CLI, run `golangci-lint run`

Optionally, to integrate golangci-lint into your editor, see [golangci-lint editor integration](https://golangci-lint.run/usage/integrations/)

### Running the tests

See [TESTS.md](https://github.com/hashicorp/terraform-provider-tfe/tree/main/TESTS.md).

### Smoke Testing Tips

After creating new schema, it's important to test your changes beyond the automated testing provided by the framework. Use these tips to ensure your provider resources behave as expected.

- Is the resource replaced when non-updatable attributes are changed?
- Is the resource unchanged after successive plans with no config changes?
- Are mutually exclusive config arguments constrained by an error?
- If adding a new argument to an existing resource: is it required? (This would be a breaking change)
- If adding a new attribute to an existing resource: is new or unexpected API authorization required?

### Updating the Changelog

Only update the `Unreleased` section. Make sure you change the unreleased tag to an appropriate version, using [Semantic Versioning](https://semver.org/) as a guideline.

Please use the template below when updating the changelog:
```
<change category>:
* **New Resource:** `name_of_new_resource` ([#123](link-to-PR))
* r/tfe_resource: description of change or bug fix ([#124](link-to-PR))
```

### Updating the documentation

For pull requests that update provider documentation, please help us verify that the
markdown will display correctly on the Registry:

- Copy the new markdown and paste it here to preview: https://registry.terraform.io/tools/doc-preview
- Paste a screenshot of that preview in your pull request.

### Change categories

- BREAKING CHANGES: Use this for any changes that aren't backwards compatible. Include details on how to handle these changes.
- FEATURES: Use this for any large new features added.
- ENHANCEMENTS: Use this for smaller new features added.
- BUG FIXES: Use this for any bugs that were fixed.
- NOTES: Use this section if you need to include any additional notes on things like upgrading, upcoming deprecations, or any other information you might want to highlight.
