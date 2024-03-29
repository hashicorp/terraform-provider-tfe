---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_default_settings
description: |-
  Sets the workspace defaults for an organization
---


<!-- Please do not edit this file, it is generated. -->
# tfe_organization_default_settings

Primarily, this is used to set the default execution mode of an organization. Settings configured here will be used as the default for all workspaces in the organization, unless they specify their own values with a [`tfeWorkspaceSettings` resource](workspace_settings.html) (or deprecated attributes on the workspace resource).

## Example Usage

Basic usage:

```typescript
import * as constructs from "constructs";
import * as cdktf from "cdktf";
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import * as tfe from "./.gen/providers/tfe";
class MyConvertedCode extends cdktf.TerraformStack {
  constructor(scope: constructs.Construct, name: string) {
    super(scope, name);
    const tfeOrganizationTest = new tfe.organization.Organization(
      this,
      "test",
      {
        email: "admin@company.com",
        name: "my-org-name",
      }
    );
    const tfeAgentPoolMyAgents = new tfe.agentPool.AgentPool(
      this,
      "my_agents",
      {
        name: "agent_smiths",
        organization: cdktf.Token.asString(tfeOrganizationTest.name),
      }
    );
    const tfeOrganizationDefaultSettingsOrgDefault =
      new tfe.organizationDefaultSettings.OrganizationDefaultSettings(
        this,
        "org_default",
        {
          defaultAgentPoolId: cdktf.Token.asString(tfeAgentPoolMyAgents.id),
          defaultExecutionMode: "agent",
          organization: cdktf.Token.asString(tfeOrganizationTest.name),
        }
      );
    new tfe.workspace.Workspace(this, "my_workspace", {
      dependsOn: [tfeOrganizationDefaultSettingsOrgDefault],
      name: "my-workspace",
    });
  }
}

```

## Argument Reference

The following arguments are supported:

* `defaultExecutionMode` - (Optional) Which [execution mode](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings#execution-mode)
  to use as the default for all workspaces in the organization. Valid values are `remote`, `local` or`agent`.
* `defaultAgentPoolId` - (Optional) The ID of an agent pool to assign to the workspace. Requires `defaultExecutionMode` to be set to `agent`. This value _must not_ be provided if `defaultExecutionMode` is set to any other value.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.


## Import

Organization default execution mode can be imported; use `<ORGANIZATION NAME>` as the import ID. For example:

```shell
terraform import tfe_organization_default_execution_mode.test my-org-name
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-131403d8bfaba1dae78b9751415a0563f3f8e7f7bc52ca8bb1517c4637beb7bb -->