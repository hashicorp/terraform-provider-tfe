---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team_notification_configuration"
description: |-
  Manages team notifications configurations.
---

# tfe_team_notification_configuration

HCP Terraform can be configured to send notifications to a team for certain events.
Team notification configurations allow you to specify a URL, destination type, and what events will trigger the notification.
Each team can have up to 20 notification configurations, and they apply to configured events for all workspaces that the configured team has access to.

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = tfe_organization.test.id
}

resource "tfe_team_notification_configuration" "test" {
  name             = "my-test-notification-configuration"
  enabled          = true
  destination_type = "generic"
  triggers         = ["change_request:created"]
  url              = "https://example.com"
  team_id          = tfe_team.test.id
}
```

With `destination_type` of `email`:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = tfe_organization.test.id
}

data "tfe_organization_membership" "test" {
  organization = tfe_organization.test.name
  email        = "example@example.com"
}

resource "tfe_team_organization_member" "test" {
  team_id                    = tfe_team.test.id
  organization_membership_id = data.tfe_organization_membership.test.id
}

resource "tfe_team_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  triggers         = ["change_request:created"]
  team_id          = tfe_team.test.id
}
```

(**TFE only**) With `destination_type` of `email`, using `email_addresses` list and `email_users`:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = tfe_organization.test.id
}

data "tfe_organization_membership" "test" {
  organization = tfe_organization.test.name
  email        = "example@example.com"
}

resource "tfe_team_organization_member" "test" {
  team_id                    = tfe_team.test.id
  organization_membership_id = data.tfe_organization_membership.test.id
}

resource "tfe_team_notification_configuration" "test" {
  name             = "my-test-email-notification-configuration"
  enabled          = true
  destination_type = "email"
  email_user_ids   = [tfe_organization_membership.test.user_id]
  email_addresses  = ["user1@company.com", "user2@company.com", "user3@company.com"]
  triggers         = ["change_request:created"]
  team_id          = tfe_team.test.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the notification configuration.
- `destination_type` - (Required) The type of notification configuration payload to send.
  Valid values are:
  - `generic`
  - `email` available in HCP Terraform or Terraform Enterprise v202005-1 or later
  - `slack`
  - `microsoft-teams` available in HCP Terraform or Terraform Enterprise v202206-1 or later
- `email_addresses` - (Optional) **TFE only** A list of email addresses. This value
  _must not_ be provided if `destination_type` is `generic`, `microsoft-teams`, or `slack`.
- `email_user_ids` - (Optional) A list of user IDs. This value _must not_ be provided
  if `destination_type` is `generic`, `microsoft-teams`, or `slack`.
- `enabled` - (Optional) Whether the notification configuration should be enabled or not.
  Disabled configurations will not send any notifications. Defaults to `false`.
- `token` - (Optional) A write-only secure token for the notification configuration, which can
  be used by the receiving server to verify request authenticity when configured for notification
  configurations with a destination type of `generic`. Defaults to `null`.
  This value _must not_ be provided if `destination_type` is `email`, `microsoft-teams`, or `slack`.
- `triggers` - (Optional) The array of triggers for which this notification configuration will
  send notifications. Currently, the only valid value is `change_request:created`.
- `url` - (Required if `destination_type` is `generic`, `microsoft-teams`, or `slack`) The HTTP or HTTPS URL of the notification configuration where notification requests will be made. This value _must not_ be provided if `destination_type` is `email`.
- `team_id` - (Required) The ID of the team that owns the notification configuration.

## Attributes Reference

- `id` - The ID of the notification configuration.

## Import

Team notification configurations can be imported; use `<NOTIFICATION CONFIGURATION ID>` as the import ID. For example:

```shell
terraform import tfe_team_notification_configuration.test nc-qV9JnKRkmtMa4zcA
```
