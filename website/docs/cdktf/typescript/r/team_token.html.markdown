---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_token"
description: |-
  Generates a new team token and overrides existing token if one exists.
---


<!-- Please do not edit this file, it is generated. -->
# tfe_team_token

Generates a new team token and overrides existing token if one exists.

## Example Usage

Basic usage:

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { Team } from "./.gen/providers/tfe/team";
import { TeamToken } from "./.gen/providers/tfe/team-token";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    const test = new Team(this, "test", {
      name: "my-team-name",
      organization: "my-org-name",
    });
    const tfeTeamTokenTest = new TeamToken(this, "test_1", {
      teamId: test.id,
    });
    /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
    tfeTeamTokenTest.overrideLogicalId("test");
  }
}

```

## Argument Reference

The following arguments are supported:

* `teamId` - (Required) ID of the team.
* `forceRegenerate` - (Optional) If set to `true`, a new token will be
  generated even if a token already exists. This will invalidate the existing
  token!
* `expiredAt` - (Optional) The token's expiration date. The expiration date must be a date/time string in RFC3339 
format (e.g., "2024-12-31T23:59:59Z"). If no expiration date is supplied, the expiration date will default to null and 
never expire.

## Example Usage

When a token has an expiry:

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { Token, TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { Team } from "./.gen/providers/tfe/team";
import { TeamToken } from "./.gen/providers/tfe/team-token";
import { Rotating } from "./.gen/providers/time/rotating";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    /*The following providers are missing schema information and might need manual adjustments to synthesize correctly: time.
    For a more precise conversion please use the --provider flag in convert.*/
    const test = new Team(this, "test", {
      name: "my-team-name",
      organization: "my-org-name",
    });
    const example = new Rotating(this, "example", {
      rotation_days: 30,
    });
    const tfeTeamTokenTest = new TeamToken(this, "test_2", {
      expiredAt: Token.asString(example.rotationRfc3339),
      teamId: test.id,
    });
    /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
    tfeTeamTokenTest.overrideLogicalId("test");
  }
}

```

## Attributes Reference

* `id` - The ID of the token.
* `token` - The generated token.

## Import

Team tokens can be imported; use `<TEAM ID>` as the import ID. For example:

```shell
terraform import tfe_team_token.test team-47qC3LmA47piVan7
```

<!-- cache-key: cdktf-0.20.8 input-81ad21e38f7d39a442070952309741b9fc85572d00ada484fd6850ada6613dff -->