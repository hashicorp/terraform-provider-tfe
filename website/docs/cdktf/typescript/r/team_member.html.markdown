---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_member"
description: |-
  Add or remove a user from a team.
---


<!-- Please do not edit this file, it is generated. -->
# tfe_team_member

Add or remove a user from a team.

~> **NOTE** on managing team memberships: Terraform currently provides four
resources for managing team memberships.
The [tfe_team_organization_member](team_organization_member.html) and [tfe_team_organization_members](team_organization_members.html) resources are
the preferred way. The [tfe_team_member](team_member.html)
resource can be used multiple times as it manages the team membership for a
single user.  The [tfe_team_members](team_members.html) resource, on the other
hand, is used to manage all team memberships for a specific team and can only be
used once. All four resources cannot be used for the same team simultaneously.

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
import { TeamMember } from "./.gen/providers/tfe/team-member";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    const test = new Team(this, "test", {
      name: "my-team-name",
      organization: "my-org-name",
    });
    const tfeTeamMemberTest = new TeamMember(this, "test_1", {
      teamId: test.id,
      username: "sander",
    });
    /*This allows the Terraform resource name to match the original name. You can remove the call if you don't need them to match.*/
    tfeTeamMemberTest.overrideLogicalId("test");
  }
}

```

## Argument Reference

The following arguments are supported:

* `teamId` - (Required) ID of the team.
* `username` - (Required) Name of the user to add.

## Import

A team member can be imported; use `<TEAM ID>/<USERNAME>` as the import ID. For
example:

```shell
terraform import tfe_team_member.test team-47qC3LmA47piVan7/sander
```

<!-- cache-key: cdktf-0.20.8 input-b59106a9c98c380491272acd9b2d6ddeddacf84931145687009cac53a30e540e -->