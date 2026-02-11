# Setup Provider to Debug Locally

Find more information [here](https://developer.hashicorp.com/terraform/plugin/debugging#starting-a-provider-in-debug-mode)

Clone the repository and build the provider binary with the necessary Go compiler flags: `-gcflags=all=-N -l`, to disable compiler optimization in order for the debugger to work efficiently.

```sh
$ git clone git@github.com:hashicorp/terraform-provider-tfe
$ cd terraform-provider-tfe
$ go build -gcflags="all=-N -l" -o {where to place the binary}
```

example, replace {platform}.
```sh
go build -gcflags="all=-N -l" -o bin/registry.terraform.io/hashicorp/tfe/9.9.9/{platform}/terraform-provider-tfe
```

You can activate the debugger via your editor such as [visual studio code](https://developer.hashicorp.com/terraform/plugin/debugging#visual-studio-code) or the Delve CLI.


#### Delve

```sh
dlv exec \
--accept-multiclient \
--continue \
--headless {location of the binary} \
-- -debug
```

example
```sh
dlv exec \
--accept-multiclient \
--continue \
--headless bin/registry.terraform.io/hashicorp/tfe/9.9.9/{platform}/terraform-provider-tfe \
-- -debug
```

##### Visual Studio Code

Example taken from [here](https://developer.hashicorp.com/terraform/plugin/debugging#visual-studio-code)
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Terraform Provider",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            // this assumes your workspace is the root of the repo
            "program": "${workspaceFolder}",
            "env": {},
            "args": [
                "-debug",
            ]
        }
    ]
}

```

You'll know you activated the debugger successfully if you see the following output.

*For vscode, the output will be located in the Debug Console tab.*

```sh
# Provider server started
export TF_REATTACH_PROVIDERS='{...}'
```

In the other project make sure you're pointing to your local provider binary you created in the previous step.

Can leverage `.terraformrc` file to override Terraform's default installation behaviors and use a local mirror for the providers you wish to use.

example:

```
provider_installation {
  filesystem_mirror {
    path = "" # path to provider binary binary
    # path = "/Users/{users}/projects/terraform-provider-tfe/bin/" macos example
    include = ["registry.terraform.io/hashicorp/tfe"]
  }
}
```

Initialize Terraform in the project you wish to debug from via `terraform init`

Should see the following output with the previous examples being used

```
Initializing provider plugins...
- Finding latest version of hashicorp/tfe...
- Installing hashicorp/tfe v9.9.9...
- Installed hashicorp/tfe v9.9.9 (unauthenticated)
```

Copy the value of `TF_REATTACH_PROVIDERS` outputted by the debugger session and either export into your shell or lead your Terraform commands setting this value:

```
TF_REATTACH_PROVIDERS='{...}' terraform {command}
```

The breakpoints you have set will halt execution and show you the current variable values.

If using the Delve CLI, include the full qualifed path to set a breakpoint.

```
(delve) b /Users/{user}/path/to/terraform-provider-tfe/tfe/resource_example.go:35
```
