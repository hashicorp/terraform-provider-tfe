---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_admin_organization_settings"
description: |-
  Manages admin settings for an organization (Terraform Enterprise Only).
---


<!-- Please do not edit this file, it is generated. -->
# tfe_admin_organization_settings

Manage admin settings for an organization. This resource requires the
use of an admin token and is for Terraform Enterprise only. See example usage for
incorporating an admin token in your provider config.

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
import { AdminOrganizationSettings } from "./.gen/providers/tfe/admin-organization-settings";
import { Organization } from "./.gen/providers/tfe/organization";
import { TfeProvider } from "./.gen/providers/tfe/provider";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    new TfeProvider(this, "tfe", {
      hostname: hostname.stringValue,
      token: token.stringValue,
    });
    const admin = new TfeProvider(this, "tfe_1", {
      alias: "admin",
      hostname: hostname.stringValue,
      token: adminToken.stringValue,
    });
    const aModuleConsumer = new Organization(this, "a-module-consumer", {
      email: "admin@company.com",
      name: "my-other-org",
    });
    const aModuleProducer = new Organization(this, "a-module-producer", {
      email: "admin@company.com",
      name: "my-org",
    });
    new AdminOrganizationSettings(this, "test-settings", {
      accessBetaTools: false,
      globalModuleSharing: false,
      moduleSharingConsumerOrganizations: [aModuleConsumer.name],
      organization: aModuleProducer.name,
      provider: "${tfe.admin}",
      workspaceLimit: 15,
    });
  }
}

```

## Argument Reference

The following arguments are supported:

* `organization` - (Optional) Name of the organization. If omitted, organization provider config must be defined.
* `accessBetaTools` - (Optional) True if the organization has access to beta tool versions.
* `workspaceLimit` - (Optional) Maximum number of workspaces for this organization. If this number is set to a value lower than the number of workspaces the organization has, it will prevent additional workspaces from being created, but existing workspaces will not be affected. If set to 0, this limit will have no effect.
* `globalModuleSharing` - (Optional) If true, modules in the organization's private module repository will be available to all other organizations. Enabling this will disable any previously configured module_sharing_consumer_organizations. Cannot be true if module_sharing_consumer_organizations is set.
* `moduleSharingConsumerOrganizations` - (Optional) A list of organization names to share modules in the organization's private module repository with. Cannot be set if global_module_sharing is true.

## Attributes Reference

* `ssoEnabled` - True if SSO is enabled in this organization

<!-- cache-key: cdktf-0.20.8 input-a01854d70d1a6325a861b17fcb030f6c8a29f8efc23c51bfd27d03211d653a5f -->