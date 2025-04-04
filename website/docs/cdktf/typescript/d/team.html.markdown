---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
description: |-
  Get information on a team.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: tfe_team

Use this data source to get information about a team.

## Example Usage

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { DataTfeTeam } from "./.gen/providers/tfe/data-tfe-team";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    new DataTfeTeam(this, "test", {
      name: "my-team-name",
      organization: "my-org-name",
    });
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the team.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the team.
* `ssoTeamId` - (Optional) The [SSO Team ID](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) of the team, if it has been defined

<!-- cache-key: cdktf-0.20.8 input-d231d33c8a4a4e5d2ef8d59dddd50d6c6faa1cb5310de9b973f9095ca67523a9 -->