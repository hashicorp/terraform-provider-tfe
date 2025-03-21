---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_github_app_installation"
description: |-
Get information on the Github App Installation.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_github_app_installation

Use this data source to get information about the Github App Installation.

## Example Usage

### Finding a Github App Installation by its installation ID

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.tfe.data_tfe_github_app_installation import DataTfeGithubAppInstallation
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        DataTfeGithubAppInstallation(self, "gha_installation",
            installation_id=12345678
        )
```

### Finding a Github App Installation by its name

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.tfe.data_tfe_github_app_installation import DataTfeGithubAppInstallation
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        DataTfeGithubAppInstallation(self, "gha_installation",
            name="github_username_or_organization"
        )
```

## Argument Reference

The following arguments are supported. At least one of `name`, `installation_id` must be set.

* `installation_id` - (Optional) ID of the Github Installation. The installation ID can be found in the URL slug when visiting the installation's configuration page, e.g `https://github.com/settings/installations/12345678`.
* `name` - (Optional) Name of the Github user or organization account that installed the app.

Must be one of: `installation_id` or `name`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The internal ID of the Github Installation. This is different from the `installation_id`.

<!-- cache-key: cdktf-0.20.8 input-5d439ec2ae1e837495b8cb500e2fcfe96d47a32f9fce3a10ffac876fe18a89dc -->