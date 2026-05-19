---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_stack_deployment"
description: |-
  Get information on a Stack deployment.
---

# Data Source: tfe_stack_deployment

Use this data source to get information about a Stack Deployment

## Example Usage

```hcl
data "tfe_stack_deployment" "staging" {
  organization = "my-example-org"
  name         = "staging"
  stack        = "example-stack"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the deployment.
* `stack` - (Required) Name of the stack containing the deployment
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Stack Deployment
* `status`: The status of the Deployment, one of "healthy", "error", "warning", "paused", or "running"
* `deployed_at`: The timestamp of the last deployment in RFC 3339 format, ex. `2024-06-12T22:01:34.956Z`,
* `
* `errors_count`: The number of plans for the deployment in the "error" state
* `warnings_count`: The number of plans for the deployment in the "warning" state.
* `paused_count`: The number of plans for the deployment in the "paused" state.
