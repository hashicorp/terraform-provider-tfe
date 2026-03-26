---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_team"
description: |-
  Get information on a team.
---

# Data Source: tfe_team

Use this data source to get information about a team.

## Example Usage

```hcl
data "tfe_team" "test" {
  name         = "my-team-name"
  organization = "my-org-name"
}
```

```hcl
data "tfe_team" "team_with_org_access" {
  name         = "my-team-name"
  organization = "my-org-name"
}

output "team_org_access" {
  value = data.tfe_team.team_with_org_access.organization_access
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the team.
* `organization` - (Optional) Name of the organization. If omitted, organization must be defined in the provider config.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the team.
* `sso_team_id` - The [SSO Team ID](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/single-sign-on#team-names-and-sso-team-ids) of the team, if it has been defined.
* `organization_access` - The team's [organization access](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#organization-permissions).

The `organization_access` block exports:

* `read_workspaces` - Allow this team to view all workspaces in this organization.
* `read_projects` - Allow this team to view all projects in this organization.
* `manage_policies` - Allow members to create, edit, read, list and delete the organization's policies.
* `manage_policy_overrides` - Allow members to override soft-mandatory policy checks.
* `delegate_policy_overrides` - When this setting is enabled for a team, its members can override failed policy evaluations on projects and workspaces they manage.
* `manage_workspaces` - Grants members the ability to view, edit, delete, and assign team access to all workspaces in this organization, as well as the ability to create new workspaces in the default project.
* `manage_vcs_settings` - Allow members to manage the organization's VCS providers and SSH keys.
* `manage_providers` - Allow members to publish and delete providers in the organization's private registry.
* `manage_modules` - Allow members to publish and delete modules in the organization's private registry.
* `manage_run_tasks` - Allow members to create, update, and delete run tasks on an organization.
* `manage_projects` - Grants members the ability to view, edit, delete, and assign team access to all projects in this organization, as well as the ability to create new workspaces in any project.
* `manage_membership` - Allow members to add and remove users from the organization, and to manage the membership of teams. This permission allows members to assign themselves to other teams.
* `manage_teams` - Grant members the ability to manage membership, as well as to create and delete teams and team tokens. This permission allows members to manage all teams, including those that they are not a part of.
* `manage_organization_access` - Grant members the ability to manage team memberships, permissions, and organization access.
* `access_secret_teams` - Allow members to access secret teams. Members will be able to view all secret teams and potentially manage them depending on their level of team permissions.
* `manage_agent_pools` - Allow members to create, update, and delete the organization's agent pools.
