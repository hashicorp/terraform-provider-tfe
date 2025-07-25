---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
description: |-
  Manages teams.
---

# tfe_team

Manages teams.

## Example Usage

Basic usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}
```

Organization Permission usage:

```hcl
resource "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
  organization_access {
    manage_vcs_settings = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the team.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.
* `visibility` - (Optional) The visibility of the team ("secret" or "organization")
* `organization_access` - (Optional) Settings for the team's [organization access](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#organization-permissions).
* `sso_team_id` - (Optional) Unique Identifier to control [team membership](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) via SAML. Defaults to `null`
* `allow_member_token_management` - (Optional) Used by Owners and users with "Manage Teams" permissions to control whether team members can manage team tokens. Defaults to `true`.

The `organization_access` block supports:

* `read_workspaces` - (Optional) Allow members to view all workspaces in this organization.
* `read_projects` - (Optional) Allow members to view all projects within the organization. Requires `read_workspaces` to be set to `true`.
* `manage_policies` - (Optional) Allows members to create, edit, and delete the organization's Sentinel policies.
* `manage_policy_overrides` - (Optional) Allows members to override soft-mandatory policy checks.
* `manage_workspaces` - (Optional) Allows members to create and administrate all workspaces within the organization.
* `manage_vcs_settings` - (Optional) Allows members to manage the organization's VCS Providers and SSH keys.
* `manage_providers` - (Optional) Allow members to publish and delete providers in the organization's private registry.
* `manage_modules` - (Optional) Allow members to publish and delete modules in the organization's private registry.
* `manage_run_tasks` - (Optional) Allow members to create, edit, and delete the organization's run tasks.
* `manage_projects` - (Optional) Allow members to create and administrate all projects within the organization. Requires `manage_workspaces` to be set to `true`.
* `manage_membership` - (Optional) Allow members to add/remove users from the organization, and to add/remove users from visible teams.
* `manage_teams` - (Optional) Allow members to create, update, and delete teams.
* `manage_organization_access` - (Optional) Allow members to update the organization access settings of teams.
* `access_secret_teams` - (Optional) Allow members access to secret teams up to the level of permissions granted by their team permissions setting.
* `manage_agent_pools` - (Optional) Allow members to create, edit, and delete agent pools within their organization.

## Attributes Reference

* `id` The ID of the team.

## Import

Teams can be imported; use `<ORGANIZATION NAME>/<TEAM ID>` or `<ORGANIZATION NAME>/<TEAM NAME>` as the import ID. For
example:

```shell
terraform import tfe_team.test my-org-name/team-uomQZysH9ou42ZYY
```
or
```shell
terraform import tfe_team.test my-org-name/my-team-name
```
