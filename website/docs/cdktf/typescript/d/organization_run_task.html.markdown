---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_organization_run_task"
description: |-
  Get information on a Run task.
---

# Data Source: tfe_organization_run_task

[Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks) allow Terraform Cloud to interact with external systems at specific points in the Terraform Cloud run lifecycle. Run tasks are reusable configurations that you can attach to any workspace in an organization.

Use this data source to get information about an [Organization Run tasks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks#creating-a-run-task).

## Example Usage

```typescript
import * as constructs from "constructs";
import * as cdktf from "cdktf";
/*Provider bindings are generated by running cdktf get.
See https://cdk.tf/provider-generation for more details.*/
import * as tfe from "./.gen/providers/tfe";
class MyConvertedCode extends cdktf.TerraformStack {
  constructor(scope: constructs.Construct, name: string) {
    super(scope, name);
    new tfe.dataTfeOrganizationRunTask.DataTfeOrganizationRunTask(
      this,
      "example",
      {
        name: "task-name",
        organization: "my-org-name",
      }
    );
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Run task.
* `organization` - (Required) Name of the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `category` - The type of task.
* `description` - A short description of the the task.
* `enabled` - Whether the task will be run.
* `id` - The ID of the task.
* `url` - URL to send a task payload.

<!-- cache-key: cdktf-0.17.0-pre.15 input-7ffa3170dbbf69fd581f515eab6eaac9c5c936b21ba712b8803b32966fbb628c -->