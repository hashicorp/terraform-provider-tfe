---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project"
description: |-
Manages projects.
---


<!-- Please do not edit this file, it is generated. -->
# tfe_project

Provides a project resource.

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
import { Organization } from "./.gen/providers/tfe/organization";
import { Project } from "./.gen/providers/tfe/project";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    const testOrganization = new Organization(this, "test-organization", {
      email: "admin@company.com",
      name: "my-org-name",
    });
    new Project(this, "test", {
      name: "projectname",
      organization: testOrganization.name,
    });
  }
}

```

With tags:

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { Organization } from "./.gen/providers/tfe/organization";
import { Project } from "./.gen/providers/tfe/project";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    const testOrganization = new Organization(this, "test-organization", {
      email: "admin@company.com",
      name: "my-org-name",
    });
    new Project(this, "test", {
      name: "projectname",
      organization: testOrganization.name,
      tags: {
        cost_center: "infrastructure",
        team: "platform",
      },
    });
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the project.
    *  TFE versions v202404-2 and earlier support between 3-36 characters
    *  TFE versions v202405-1 and later support between 3-40 characters
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `description` - (Optional) A description for the project.
* `autoDestroyActivityDuration` - A duration string for all workspaces in the project, representing time after each workspace's activity when an auto-destroy run will be triggered.
* `tags` - (Optional) A map of key-value tags to add to the project.
* `ignoreAdditionalTags` - (Optional) Explicitly ignores `tags`
_not_ defined by config so they will not be overwritten by the configured
tags. This creates exceptional behavior in Terraform with respect
to `tags` and is not recommended. This value must be applied before it
will be used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The project ID.

## Import

Projects can be imported; use `<PROJECT ID>` as the import ID. For example:

```shell
terraform import tfe_project.test prj-niVoeESBXT8ZREhr
```

<!-- cache-key: cdktf-0.20.8 input-9351973607d3e9ae014783a996eb54fad1d9b32f3e65eeec985b53836806ea20 -->