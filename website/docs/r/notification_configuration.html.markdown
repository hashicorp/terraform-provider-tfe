---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_notification_configuration"
description: |-
  Manages notifications configurations.
---

# tfe_notification_configuration

HCP Terraform can be configured to send notifications for run state transitions.
Notification configurations allow you to specify a URL, destination type, and what events will trigger the notification.
Each workspace can have up to 20 notification configurations, and they apply to all runs for that workspace.

~> **NOTE:** The `url_wo` and `token_wo` arguments are write-only alternatives to `url` and `token` that are never stored in Terraform state. They are recommended over their plaintext equivalents. Write-only arguments require Terraform 1.11.0 or later. [Learn more](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments).

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  enabled          = true
  destination_type = "generic"
  triggers         = ["run:created", "run:planning", "run:errored"]
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.test.id
}
```

With `destination_type` of `email`:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "test.member@company.com"
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  triggers         = ["run:created", "run:planning", "run:errored"]
  workspace_id     = tfe_workspace.test.id
}
```

(**TFE only**) With `destination_type` of `email`, using `email_addresses` list and `email_users`:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_organization_membership" "test" {
  organization = "my-org-name"
  email        = "test.member@company.com"
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  email_addresses  = ["user1@company.com", "user2@company.com", "user3@company.com"]
  triggers         = ["run:created", "run:planning", "run:errored"]
  workspace_id     = tfe_workspace.test.id
}
```

With write-only token and URL (auto-managed, recommended):

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_workspace" "test" {
  name         = "my-workspace-name"
  organization = tfe_organization.test.id
}

resource "tfe_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  destination_type = "generic"
  token_wo         = "my-secret-token"
  url_wo           = "https://example.com"
  workspace_id     = tfe_workspace.test.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the notification configuration.
* `destination_type` - (Required) The type of notification configuration payload to send.
  Valid values are:
  * `generic`
  * `email` available in HCP Terraform or Terraform Enterprise v202005-1 or later
  * `slack`
  * `microsoft-teams` available in HCP Terraform or Terraform Enterprise v202206-1 or later
* `email_addresses` - (Optional) **TFE only** A list of email addresses. This value
  _must not_ be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.
* `email_user_ids` - (Optional) A list of user IDs. This value _must not_ be provided
  if `destination_type` is `generic`, `microsoft-teams`, or `slack`.
* `enabled` - (Optional) Whether the notification configuration should be enabled or not.
  Disabled configurations will not send any notifications. Defaults to `false`.
* `token` - (Optional) A token for the notification configuration, which can be used by the
  receiving server to verify request authenticity when configured for notification configurations
  with a destination type of `generic`. Defaults to `null`. This value _must not_ be provided
  if `destination_type` is `email`, `microsoft-teams`, or `slack`. Cannot be used with `token_wo`.
  Prefer `token_wo` to prevent the token from being stored in state.
* `token_wo` - (Optional, [Write-Only](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments))
  Write-only alternative to `token`. Never stored in Terraform state. Cannot be used with `token`.
  This value _must not_ be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.

  The provider automatically detects changes by storing a SHA-256 hash of the value in
  [private state](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state)
  and incrementing `token_wo_version` when it changes. No additional configuration is required.

  For maximum privacy — to prevent even the hash from being stored — omit `token_wo` from
  your config and set `token_wo_version` manually instead, incrementing it whenever you
  need to push a new token value.

* `token_wo_version` - (Optional) Tracks the version of `token_wo`. In **auto-managed mode**
  (the default when `token_wo_version` is not set in config), the provider computes this value
  automatically: it is set to `1` on resource creation and incremented whenever the value of
  `token_wo` changes. In **manual mode** (when you explicitly set `token_wo_version` in config),
  auto-detection is disabled and you control updates by incrementing this value yourself —
  no hash is stored in private state. Cannot be used with `token`.
* `triggers` - (Optional) The array of triggers for which this notification configuration will
  send notifications. Valid values are `run:created`, `run:planning`, `run:needs_attention`, `run:applying`
  `run:completed`, `run:errored`, `assessment:check_failure`, `assessment:drifted`, `assessment:failed`,
  `workspace:auto_destroy_reminder`, or `workspace:auto_destroy_run_results`.
  If omitted, no notification triggers are configured.
* `url` - (Optional) The HTTP or HTTPS URL where notification requests will be made. Required
  when `destination_type` is `generic`, `microsoft-teams`, or `slack` and `url_wo` is not set.
  This value _must not_ be provided if `destination_type` is `email`. Cannot be used with `url_wo`.
  Prefer `url_wo` to prevent the URL from being stored in state.
* `url_wo` - (Optional, [Write-Only](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments))
  Write-only alternative to `url`. Never stored in Terraform state. Required when
  `destination_type` is `generic`, `microsoft-teams`, or `slack` and `url` is not set.
  Cannot be used with `url`. This value _must not_ be provided if `destination_type` is `email`.

  The provider automatically detects changes by storing a SHA-256 hash of the value in
  [private state](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state)
  and incrementing `url_wo_version` when it changes. No additional configuration is required.

  For maximum privacy — to prevent even the hash from being stored — omit `url_wo` from
  your config and set `url_wo_version` manually instead, incrementing it whenever you
  need to push a new URL value.

* `url_wo_version` - (Optional) Tracks the version of `url_wo`. In **auto-managed mode**
  (the default when `url_wo_version` is not set in config), the provider computes this value
  automatically: it is set to `1` on resource creation and incremented whenever the value of
  `url_wo` changes. In **manual mode** (when you explicitly set `url_wo_version` in config),
  auto-detection is disabled and you control updates by incrementing this value yourself —
  no hash is stored in private state. Cannot be used with `url`.
* `workspace_id` - (Required) The id of the workspace that owns the notification configuration.

## Attributes Reference

* `id` - The ID of the notification configuration.

## Import

Notification configurations can be imported; use `<NOTIFICATION CONFIGURATION ID>` as the import ID. For example:

```shell
terraform import tfe_notification_configuration.test nc-qV9JnKRkmtMa4zcA
```
