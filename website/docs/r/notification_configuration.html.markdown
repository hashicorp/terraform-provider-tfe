---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_notification_configuration"
sidebar_current: "docs-resource-tfe-notification-configuration"
description: |-
  Manages notifications configurations.
---

# tfe_notification_configuration

Terraform Cloud can be configured to send notifications for run state transitions. 
Notification configurations allow you to specify a URL, destination type, and what events will trigger the notification. 
Each workspace can have up to 20 notification configurations, and they apply to all runs for that workspace.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = "${tfe_organization.test.id}"
}

resource "tfe_notification_configuration" "test" {
  name                      = "my-test-notification-configuration"
  enabled                   = true
  destination_type          = "generic"
  triggers                  = ["run:created", "run:planning", "run:errored"]
  url                       = "https://example.com"
  workspace_external_id     = "${tfe_workspace.test.external_id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the notification configuration.
* `destination_type` - (Required) The type of notification configuration payload to send. 
  Valid values are `generic` or `slack`.
* `enabled` - (Optional) Whether the notification configuration should be enabled or not.
  Disabled configurations will not send any notifications. Defaults to `false`.
* `token` - (Optional) A write-only secure token for the notification configuration, which can
  be used by the receiving server to verify request authenticity when configured for notification
  configurations with a destination type of `generic`. A token set for notification configurations
  with a destination type of `slack` is not allowed and will result in an error. Defaults to `null`.
* `triggers` - (Optional) The array of triggers for which this notification configuration will
  send notifications. Valid values are `run:created`, `run:planning`, `run:needs_attention`, `run:applying`
  `run:completed`, `run:errored`. If omitted, no notification triggers are configured.
* `url` - (Required) The HTTP or HTTPS URL of the notification configuration where notification
  requests will be made.
* `workspace_external_id` - (Required) The external id of the workspace that owns the notification configuration.

## Attributes Reference

* `id` - The ID of the notification configuration.

## Import

Notification configurations can be imported; use `<NOTIFICATION CONFIGURATION ID>` as the import ID. For example:

```shell
terraform import tfe_notification_configuration.test nc-qV9JnKRkmtMa4zcA
```
