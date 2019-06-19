Terraform Enterprise Provider
=============================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Building The Provider
---------------------

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

Using the provider
----------------------
If you're building the provider, follow the instructions to
[install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin)
After placing it into your plugins directory,  run `terraform init` to initialize it.

Developing the Provider
---------------------------

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

Testing
-------

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

```sh
$ make testacc
```

A hostname and token must be provided in order to run the acceptance tests. By
default, these are loaded from the the `credentials` in the [CLI config
file](https://www.terraform.io/docs/commands/cli-config.html). You can override
these values with the environment variables specified below: `TFE_HOSTNAME` and
`TFE_TOKEN`.

To run all tests, you will need to set the following environment variables:

- `GITHUB_TOKEN`: a GitHub personal access token, used to establish a VCS provider connection
- `TFE_HOSTNAME`: the hostname of your test TFE instance; for example, `tfe-test.local`
- `TFE_POLICY_SET_VCS_BRANCH`: a VCS branch, used to test policy sets
- `TFE_POLICY_SET_VCS_PATH`: a VCS path, used to test policy sets
- `TFE_TOKEN`: a user token for an administrator account on your TFE instance
- `TFE_USER1` and `TFE_USER2`: the usernames of two pre-existing TFE users, for testing team membership
- `TFE_VCS_IDENTIFIER`: a VCS identifier, used to test policy sets
