## Unreleased

BUG FIXES:
* `r/tfe_team_project_access`: Fixes a panic that occurs when the client is configured against an older TFE release, by @sebasslash [1011](https://github.com/hashicorp/terraform-provider-tfe/pull/1011)

FEATURES:
* `d/tfe_organization_membership`: Add `organization_membership_id` attribute, by @laurenolivia [997](https://github.com/hashicorp/terraform-provider-tfe/pull/997)
* `d/tfe_variable_set`: Add `project_ids` attribute, by @Netra2104 [994](https://github.com/hashicorp/terraform-provider-tfe/pull/994)
* **New Data Source**: `d/tfe_teams` is a new data source to return names and IDs of Teams in an Organization, by @isaacmcollins [992](https://github.com/hashicorp/terraform-provider-tfe/pull/992)

## v0.48.0 (August 7, 2023)

BUG FIXES:
* `r/tfe_workspace`: Fix panic when updating `trigger_patterns` attribute, by @liamstevens [969](https://github.com/hashicorp/terraform-provider-tfe/pull/969)
* `r/tfe_admin_organization_settings`: Allow reprovisioning when the parent organization has been deleted, by @ctrombley [982](https://github.com/hashicorp/terraform-provider-tfe/pull/982)

FEATURES:
* **New Resource**: `r/tfe_saml_settings` manages SAML Settings, by @karvounis-form3 [970](https://github.com/hashicorp/terraform-provider-tfe/pull/970)
* `d/tfe_saml_settings`: Add PrivateKey (sensitive), SignatureSigningMethod, and SignatureDigestMethod attributes, by @karvounis-form3 [970](https://github.com/hashicorp/terraform-provider-tfe/pull/970)
* **New Resource**: `r/tfe_project_policy_set` is a new resource to attach/detach an existing `project` to an existing `policy set`, by @Netra2104 [972](https://github.com/hashicorp/terraform-provider-tfe/pull/972)
* `d/tfe_policy_set`: Add `project_ids` attribute, by @Netra2104 [974](https://github.com/hashicorp/terraform-provider-tfe/pull/974/files)
* `r/tfe_team_project_access`: Add a `custom` option to the `access` attribute as well as `project_access` and `workspace_access` attributes with
various customizable permissions options to apply to a project and all of the workspaces therein, by @rberecka [983](https://github.com/hashicorp/terraform-provider-tfe/pull/983)
* `d/team_project_access`: Add a `custom` option to the `access` attribute as well as `project_access` and `workspace_access` attributes, by @rberecka [983](https://github.com/hashicorp/terraform-provider-tfe/pull/983)


NOTES:
* The provider is now using go-tfe [v1.32.0](https://github.com/hashicorp/go-tfe/releases/tag/v1.32.0)
## v0.47.0 (July 18, 2023)

FEATURES:
* **New Data Source**: `d/tfe_saml_settings` is a new data source to retrieve SAML settings from the Terraform Enterprise Admin API, by @karvounis-form3 [952](https://github.com/hashicorp/terraform-provider-tfe/pull/952)

BUG FIXES:
* `d/tfe_project`: Ignore case when matching project name from Projects List API, by @jbonhag [958](https://github.com/hashicorp/terraform-provider-tfe/pull/958)

## v0.46.0 (July 3, 2023)

FEATURES:
* **New Resource**: `r/tfe_agent_pool_allowed_workspaces` restricts the use of an agent pool to particular workspaces, by @hs26gill [870](https://github.com/hashicorp/terraform-provider-tfe/pull/870)
* `r/tfe_organization_token`: Add optional `expired_at` field to organization tokens, by @juliannatetreault ([#844](https://github.com/hashicorp/terraform-provider-tfe/pull/844))
* `r/tfe_team_token`: Add optional `expired_at` field to team tokens, by @juliannatetreault ([#844](https://github.com/hashicorp/terraform-provider-tfe/pull/844))
* `r/tfe_agent_pool`: Add attribute `organization_scoped` to set the scope of an agent pool, by @hs26gill [870](https://github.com/hashicorp/terraform-provider-tfe/pull/870)
* `d/tfe_agent_pool`: Add attribute `organization_scoped` and `allowed_workspace_ids` to retrieve agent pool scope and associated allowed workspace ids, by @hs26gill [870](https://github.com/hashicorp/terraform-provider-tfe/pull/870)

BUG FIXES:
* `r/tfe_workspace_run`: Ensure `wait_for_run` correctly results in a fire-and-forget run when set to `false`, by @lucymhdavies ([#910](https://github.com/hashicorp/terraform-provider-tfe/pull/910))
* `r/tfe_workspace_run`: Fix rare random run failures; adjust lists of expected run statuses to ensure that a plan is completely processed before attempting to apply it, by @uk1288 ([#921](https://github.com/hashicorp/terraform-provider-tfe/pull/921))
* `r/tfe_notification_configuration`: Add support for missing "Check failed" Health Event notifications, by @lucymhdavies ([#927](https://github.com/hashicorp/terraform-provider-tfe/pull/927))
* `r/tfe_registry_module`: Fix a bug that prevented users from being able to create a registry module using a github app, by @dsa0x ([#935](https://github.com/hashicorp/terraform-provider-tfe/pull/935))

## v0.45.0 (May 25, 2023)

FEATURES:
* `r/tfe_team`: Add attribute `manage_membership` to `organization_access` on `tfe_team` by @JarrettSpiker ([#801](https://github.com/hashicorp/terraform-provider-tfe/pull/801))
* **New Resource**: `r/tfe_workspace_run` manages create and destroy lifecycles in a workspace, by @uk1288 ([#786](https://github.com/hashicorp/terraform-provider-tfe/pull/786))
* `r/tfe_variable`: Add a `readable_value` attribute, which will provide an un-redacted representation of the variable's value in plan outputs if the variable is not sensitive, and which may be referenced by downstream resources by @JarrettSpiker ([#801](https://github.com/hashicorp/terraform-provider-tfe/pull/867))

ENHANCEMENTS:
* `r/tfe_workspace`: Retry workspace safe delete if resources are still being processed to determine safety. ([#881](https://github.com/hashicorp/terraform-provider-tfe/pull/881))

BUG FIXES:

* `r/tfe_variable`: Don't silently erase or override the `value` of a sensitive variable on changes to other attributes when `ignore_changes = [value]` is set, by @nfagerlund ([#873](https://github.com/hashicorp/terraform-provider-tfe/pull/873), fixing issue [#839](https://github.com/hashicorp/terraform-provider-tfe/issues/839))

## v0.44.1 (April 21, 2023)

BUG FIXES:

* Fixed a documentation bug in the new `r/tfe_no_code_module` resource, incorrectly labelling the attribute `registry_module` as `module`

## v0.44.0 (April 19, 2023)

FEATURES:
* **New Data Source**: `d/tfe_project` is a new data source to retrieve project id and associated workspace ids, by @hs26gill ([#829](https://github.com/hashicorp/terraform-provider-tfe/pull/829))
* **New Resource**: `r/tfe_project_variable_set` is a new resource to apply variable sets to projects, by @jbonhag and @rberecka ([#837](https://github.com/hashicorp/terraform-provider-tfe/pull/837))
* **New Resource**: `r/tfe_no_code_module` is a new resource to manage no-code settings for registry modules, by @dsa0x ([#836](https://github.com/hashicorp/terraform-provider-tfe/pull/836))

    **NOTE:** This resource is currently in beta and isn't generally available to all users. It is subject to change or removal.

BUG FIXES:
* `r/tfe_workspace`: Only set `oauth_token_id` and `github_app_installation_id` if configured, by @moensch ([#835](https://github.com/hashicorp/terraform-provider-tfe/pull/835))

DEPRECATIONS:

* The `no_code` attribute in r/tfe_registry_module is deprecated in favor of the new resource `tfe_no_code_module`, which provides a more flexible interface for managing no-code settings for registry modules. The `no_code` attribute will be removed in the next major release of the provider. By @dsa0x ([#836](https://github.com/hashicorp/terraform-provider-tfe/pull/836))

## v0.43.0 (March 23, 2023)

FEATURES:
* **New Data Source**: `d/tfe_organization_tags` is a new data source to allow reading all workspace tags within an organization, by @rhughes1 ([#773](https://github.com/hashicorp/terraform-provider-tfe/pull/773))
* **New Data Source**: `d/tfe_github_app_installation` is a new data source to read a github app installation by name or github app in installation id, by @roleesinhaHC ([#808](https://github.com/hashicorp/terraform-provider-tfe/pull/808))
* `r/tfe_workspace`: Add attribute `github_app_installation_id` to the `vcs_repo`, by @roleesinhaHC ([#808](https://github.com/hashicorp/terraform-provider-tfe/pull/808))
* `r/tfe_registry_module`: Add attribute `github_app_installation_id` to the `vcs_repo`, by @roleesinhaHC ([#808](https://github.com/hashicorp/terraform-provider-tfe/pull/808))
* `r/tfe_policy_set`: Add attribute `github_app_installation_id` to the `vcs_repo`, by @roleesinhaHC ([#808](https://github.com/hashicorp/terraform-provider-tfe/pull/808))
* `r/tfe_workspace`, `d/tfe_workspace`: Add `source_name` and `source_url` to workspaces, by @lucymhdavies ([#527](https://github.com/hashicorp/terraform-provider-tfe/pull/527))
* `r/tfe_team`: Add `read_projects` and `read_workspaces` to the `organization_access` block, by @SwiftEngineer ([#796](https://github.com/hashicorp/terraform-provider-tfe/pull/796))
* `r/tfe_team_project_access` and `d/tfe_team_project_access`: Added support for "maintain" and "write" project permissions, by @joekarl and @jbonhag ([#826](https://github.com/hashicorp/terraform-provider-tfe/pull/826))
* `r/tfe_workspace` and `d/tfe_workspace`: Add attribute `html_url`, by @brandonc ([#784](https://github.com/hashicorp/terraform-provider-tfe/pull/784))
* `r/tfe_organization_membership`: Organization Memberships can now be imported using `<ORGANIZATION NAME>/<USER EMAIL>`, by @JarrettSpiker ([#715](https://github.com/hashicorp/terraform-provider-tfe/pull/715))

ENHANCEMENTS:
* Clarify usage of `organization` fields in documentation describing VCS repository config blocks, by @brandonc ([#792](https://github.com/hashicorp/terraform-provider-tfe/pull/792))
* `r/tfe_workspace`: Clarify error message shown when attempting to safe-delete a workspace on a version of TFE which does not support safe delete, by @JarrettSpiker ([#803](https://github.com/hashicorp/terraform-provider-tfe/pull/803))

## v0.42.0 (January 31, 2023)

FEATURES:
* **New Provider Config**: `organization` (or the `TFE_ORGANIZATION` environment variable) defines a default organization for all resources, making all resource-specific organization arguments optional, by @brandonc ([#762](https://github.com/hashicorp/terraform-provider-tfe/pull/762))
* **New Resource**: `r/tfe_team_project_access` manages team project permissions, by @mwudka ([#768](https://github.com/hashicorp/terraform-provider-tfe/pull/768))
* **New Data Source**: `d/tfe_team_project_access` reads existing team project permissions, by @mwudka ([#768](https://github.com/hashicorp/terraform-provider-tfe/pull/768))
* `r/tfe_team`: Add attribute `manage_projects` to `tfe_team`, by @mwudka ([#768](https://github.com/hashicorp/terraform-provider-tfe/pull/768))
* `r/tfe_team`: Teams can now be imported using `<ORGANIZATION NAME>/<TEAM NAME>`, by @JarrettSpiker ([#745](https://github.com/hashicorp/terraform-provider-tfe/pull/745))
* `r/tfe_team_organization_member`: Team Organization Memberships can now be imported using `<ORGANIZATION NAME>/<USER EMAIL>/<TEAM NAME>`, by @JarrettSpiker ([#745](https://github.com/hashicorp/terraform-provider-tfe/pull/745))

ENHANCEMENTS:
* Update API doc links from terraform.io to developer.hashicorp domain by @uk1288 [#764](https://github.com/hashicorp/terraform-provider-tfe/pull/764)
* Update website docs to depict the use of set with `tfe_team_organization_members` and `tfe_team_members` by @uk1288 [#767](https://github.com/hashicorp/terraform-provider-tfe/pull/767)
* `d/tfe_workspace`: Add `execution_mode` field to workspace datasource @Uk1288 ([#772](https://github.com/hashicorp/terraform-provider-tfe/pull/772))

BUG FIXES:
* `r/tfe_workspace`: Return all workspace safe deletion errors by @skeggse ([#758](https://github.com/hashicorp/terraform-provider-tfe/pull/758))

## v0.41.0 (January 4, 2023)

BUG FIXES:
* d/tfe_workspace_ids: When no wildcards were used in the names argument a substring match was being performed anyway @brandonc ([#752](https://github.com/hashicorp/terraform-provider-tfe/pull/752))

FEATURES:
* r/tfe_workspace: Add attribute `resource_count` to `tfe_workspace` by @rhughes1 ([#682](https://github.com/hashicorp/terraform-provider-tfe/pull/682))
* d/tfe_outputs: Add `nonsensitive_values` attribute to expose current non-sensitive outputs of a given workspace @Uk1288 ([#711](https://github.com/hashicorp/terraform-provider-tfe/pull/711))
* r/tfe_workspace: Adds validation to tag_names argument to ensure tags are lowercase and don't contain invalid characters @brandonc ([#743](https://github.com/hashicorp/terraform-provider-tfe/pull/743))

## v0.40.0 (December 6, 2022)

DEPRECATIONS:
* r/tfe_sentinel_policy is deprecated in favor of the new resource `tfe_policy`, which supports both Sentinel and OPA policies
* r/tfe_organization_module_sharing is deprecated in favor of the new resource `tfe_admin_organization_settings`, which supports the global module sharing option

FEATURES:
* **New Resource**: `tfe_admin_organization_settings` ([#709](https://github.com/hashicorp/terraform-provider-tfe/pull/709)) adds the ability for Terraform Enterprise admins to configure settings for an organization, including module consumers and global module sharing config.
* **New Resource**: `tfe_policy` is a new resource that supports both Sentinel as well as OPA policies. `tfe_sentinel_policy` now includes a deprecation warning. ([#690](https://github.com/hashicorp/terraform-provider-tfe/pull/690))
* **New Resource**: `tfe_project` allows managing projects, which is an upcoming feature of Terraform Cloud and may not yet be generally available. ([#704](https://github.com/hashicorp/terraform-provider-tfe/pull/704))
* d/tfe_workspace_ids: Add support for filtering workspace names with partial matching using `*` ([#698](https://github.com/hashicorp/terraform-provider-tfe/pull/698))
* r/tfe_workspace: Add preemptive check for resources under management when `force_delete` attribute is false ([#699](https://github.com/hashicorp/terraform-provider-tfe/pull/699))
* r/tfe_policy_set: Add OPA support for policy sets. ([#691](https://github.com/hashicorp/terraform-provider-tfe/pull/691))
* d/tfe_policy_set: Add optional `kind` and `overridable` fields for OPA policy sets ([#691](https://github.com/hashicorp/terraform-provider-tfe/pull/691))
* r/tfe_policy: enforce_mode is no longer a required property ([#705](https://github.com/hashicorp/terraform-provider-tfe/pull/705))
* d/tfe_organization: Add computed `default_project_id` field to support projects ([#704](https://github.com/hashicorp/terraform-provider-tfe/pull/704))
* r/tfe_workspace: Add optional `project_id` argument to support projects ([#704](https://github.com/hashicorp/terraform-provider-tfe/pull/704))
* d/tfe_workspace: Add optional `project_id` argument to support projects ([#704](https://github.com/hashicorp/terraform-provider-tfe/pull/704))

## v0.39.0 (November 18, 2022)

FEATURES:
* r/tfe_workspace_run_task: Removed beta notices on the `stage` attribute for workspace run tasks. ([#669](https://github.com/hashicorp/terraform-provider-tfe/pull/669))
* r/registry_module: Adds `no_code` field. ([#673](https://github.com/hashicorp/terraform-provider-tfe/pull/673))
* r/tfe_organization: Add `allow_force_delete_workspaces` attribute to set whether admins are permitted to delete workspaces with resource under management. ([#661](https://github.com/hashicorp/terraform-provider-tfe/pull/661))
* r/tfe_workspace: Add `force_delete` attribute to set whether workspaces will be force deleted when removed through the provider. Otherwise, they will be safe deleted. ([#675](https://github.com/hashicorp/terraform-provider-tfe/pull/675))
* r/tfe_notification_configuration: Add assessment triggers to notifications ([#676](https://github.com/hashicorp/terraform-provider-tfe/pull/676))

## v0.38.0 (October 24, 2022)

FEATURES:
* d/tfe_oauth_client: Adds `name`, `service_provider`, `service_provider_display_name`, `organization`, `callback_url`, and `created_at` fields, and enables searching for an OAuth client with `organization`, `name`, and `service_provider`. ([#599](https://github.com/hashicorp/terraform-provider-tfe/pull/599))
* d/tfe_organization_members: Add datasource for organization_members that returns a list of active members and members with pending invite in an organization. ([#635](https://github.com/hashicorp/terraform-provider-tfe/pull/635))
* d/tfe_organization_membership: Add new argument `username` to enable fetching an organization membership by username. ([#660](https://github.com/hashicorp/terraform-provider-tfe/pull/660))
* r/tfe_organization_membership: Add new computed attribute `username`. ([#660](https://github.com/hashicorp/terraform-provider-tfe/pull/660))
* r/tfe_team_organization_members: Add resource for managing team members via organization membership IDs ([#617](https://github.com/hashicorp/terraform-provider-tfe/pull/617))

BUG FIXES:
* r/tfe_workspace: When assessments_enabled was the only change in to the resource the workspace was not being updated ([#641](https://github.com/hashicorp/terraform-provider-tfe/pull/641))

NOTES:
* The provider is now using go 1.18. ([#643](https://github.com/hashicorp/terraform-provider-tfe/pull/643), [#646](https://github.com/hashicorp/terraform-provider-tfe/pull/646))

## v0.37.0 (September 28, 2022)

FEATURES:
* r/tfe_workspace: Changes in `agent_pool_id` and `execution_mode` attributes are now detected and applied. ([#607](https://github.com/hashicorp/terraform-provider-tfe/pull/607))
* r/tfe_workspace_run_task, d/tfe_workspace_run_task: Add `stage` attribute to workspace run tasks. ([#555](https://github.com/hashicorp/terraform-provider-tfe/pull/555))
* r/tfe_workspace_policy_set: Add ability to attach an existing `workspace` to an existing `policy set`. ([#591](https://github.com/hashicorp/terraform-provider-tfe/pull/591))
* Add attributes for health assessments (drift detection) - available only in Terraform Cloud ([550](https://github.com/hashicorp/terraform-provider-tfe/pull/550)):
  * r/tfe_workspace: Add attribute `assessments_enabled`
  * d/tfe_workspace: Add attribute `assessments_enabled`
  * r/tfe_organization: Added attribute `assessments_enforced`
  * d/tfe_organization: Added attribute `assessments_enforced`

BUG FIXES:
* Bump `terraform-plugin-go` to `v0.6.0`, due to a crash when `tfe_outputs` had null values. ([#611](https://github.com/hashicorp/terraform-provider-tfe/pull/611))
* r/tfe_workspace: Fix documentation of file_triggers_enabled default. ([#627](https://github.com/hashicorp/terraform-provider-tfe/pull/627))
* r/tfe_variable_set: Fix panic when applying variable set to workspaces fails ([#628](https://github.com/hashicorp/terraform-provider-tfe/pull/628))

## v0.36.0 (August 16th, 2022)

FEATURES:
* r/tfe_organization_run_task, d/tfe_organization_run_task: Add `description` attribute to organization run tasks. ([#585](https://github.com/hashicorp/terraform-provider-tfe/pull/585))
* d/tfe_policy_set: Add datasource for policy_set ([#592](https://github.com/hashicorp/terraform-provider-tfe/pull/592))
* r/tfe_workspace: Adds `tags_regex` attribute to `vcs_repo` for workspaces, enabling a workspace to trigger runs for matching Git tags. ([#549](https://github.com/hashicorp/terraform-provider-tfe/pull/549))
* r/agent_pool: Agent Pools can now be imported using `<ORGANIZATION NAME>/<AGENT POOL NAME>` ([#561](https://github.com/hashicorp/terraform-provider-tfe/pull/561))

BUG FIXES:
* d/tfe_outputs: Fix a bug causing sensitive values to be missing from tfe_outputs ([#565](https://github.com/hashicorp/terraform-provider-tfe/pull/565))

## 0.35.0 (July 27th, 2022)

BREAKING CHANGES:
* `r/tfe_organization`: `admin_settings` attribute was removed after being released prematurely in 0.34.0, breaking existing configurations due to requiring a token with admin privileges ([#573](https://github.com/hashicorp/terraform-provider-tfe/pull/573))

BUG FIXES:
* r/tfe_registry_module: Added `Computed` modifier to attributes in order to prevent unnecessary resource replacement ([#572](https://github.com/hashicorp/terraform-provider-tfe/pull/572))

## 0.34.0 (July 26th, 2022)

BUG FIXES:
* Removed nonworking example from `tfe_variable_set` docs ([#562](https://github.com/hashicorp/terraform-provider-tfe/pull/562))
* Removed `ForceNew` modifier from `name` attribute in `r/tfe_team` ([#566](https://github.com/hashicorp/terraform-provider-tfe/pull/566))
* r/tfe_workspace: Fix `trigger-prefixes` could not be updated because of the conflict with `trigger-patterns` in some cases - as described in this [GitHub Issue](https://github.com/hashicorp/terraform-provider-tfe/issues/552) ([#564](https://github.com/hashicorp/terraform-provider-tfe/pull/564/))

FEATURES:
* d/agent_pool: Improve efficiency of reading agent pool data when the target organization has more than 20 agent pools ([#508](https://github.com/hashicorp/terraform-provider-tfe/pull/508))
* Added warning logs for 404 error responses ([#538](https://github.com/hashicorp/terraform-provider-tfe/pull/538))
* r/tfe_registry_module: Add ability to create both public and private `registry_modules` without VCS. ([#546](https://github.com/hashicorp/terraform-provider-tfe/pull/546))

DEPRECATION NOTICE:
* The `registry_modules` import format `<ORGANIZATION>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>` has been deprecated in favour of `<ORGANIZATION>/<REGISTRY_NAME>/<NAMESPACE>/<REGISTRY MODULE NAME>/<REGISTRY MODULE PROVIDER>/<REGISTRY MODULE ID>` to support public and private `registry_modules`.

## 0.33.0 (July 8th, 2022)

FEATURES:
* **New Resource**: `tfe_workspace_variable_set` ([#537](https://github.com/hashicorp/terraform-provider-tfe/pull/537)) adds the ability to assign a variable set to a workspace in a single, flexible resource.
* r/tfe_workspace, d/tfe_workspace: `trigger-patterns` ([#502](https://github.com/hashicorp/terraform-provider-tfe/pull/502)) attribute is introduced to support specifying a set of [glob patterns](https://www.terraform.io/cloud-docs/workspaces/settings/vcs#glob-patterns-for-automatic-run-triggering) for automatic VCS run triggering.
* r/organization: Add `workspace_limit` setting, available only in Terraform Enterprise ([#521](https://github.com/hashicorp/terraform-provider-tfe/pull/521))

DEPRECATION NOTICE: The `workspace_ids` argument on `tfe_variable_set` has been labelled as deprecated and should not be used in conjunction with `tfe_workspace_variable_set`.

## 0.32.1 (June 21st, 2022)

BUG FIXES:

* Fixed a bug in the latest release where a team data source could be populated with the wrong team. ([#530](https://github.com/hashicorp/terraform-provider-tfe/pull/530))

## 0.32.0 (June 20th, 2022)

0.32.0 is an impactful release that includes several bug fixes, support for [run tasks](https://www.terraform.io/cloud-docs/workspaces/settings/run-tasks#run-tasks) and several breaking changes that you should review carefully.

BREAKING CHANGES:
* **Removed Authentication Method**: Host-specific TF_TOKEN_... environment variable (added in 0.31.0) can no longer be used for token authentication. This method of authentication is incompatible with the Terraform Cloud remote execution model. Please use the TFE_TOKEN environment variable.
* r/tfe_workspace: Default value of the `file_triggers_enabled` field is changed to `false`. This will align the
  `file_triggers_enabled` field default value with the default value for the same field in the
  [TFC API](https://www.terraform.io/cloud-docs/api-docs/workspaces).
  If the value of the `file_triggers_enabled` field was not explicitly set and either of the fields `working_directory`
  (not an empty string) or `trigger_prefixes` was used - to keep the behavior unchanged, the `file_trigger_enabled`
  field should now explicitly be set to `true`. ([#510](https://github.com/hashicorp/terraform-provider-tfe/pull/510/files))
* r/tfe_team_access: The `permissions` attribute requires `run_tasks` in the block. ([#487](https://github.com/hashicorp/terraform-provider-tfe/pull/487))

BUG FIXES:
* Prevent overwriting `vcs_repo` attributes in `r/tfe_workspace` when update API call fails ([#498](https://github.com/hashicorp/terraform-provider-tfe/pull/498))
* Fix panic crash on `trigger_prefixes` update in `r/tfe_workspace` when given empty strings ([#518](https://github.com/hashicorp/terraform-provider-tfe/pull/518))

FEATURES:
* r/team, d/team: Add manage_run_tasks to the tfe_team organization_access attributes ([#486](https://github.com/hashicorp/terraform-provider-tfe/pull/486))
* **New Resource**: `tfe_organization_run_task` ([#488](https://github.com/hashicorp/terraform-provider-tfe/pull/488))
* **New Resource**: `tfe_workspace_run_task` ([#488](https://github.com/hashicorp/terraform-provider-tfe/pull/488))
* **New Data Source**: d/tfe_organization_run_task ([#488](https://github.com/hashicorp/terraform-provider-tfe/pull/488))
* **New Data Source**: d/tfe_workspace_run_task ([#488](https://github.com/hashicorp/terraform-provider-tfe/pull/488))
* r/tfe_notification_configuration: Add Microsoft Teams notification type ([#484](https://github.com/hashicorp/terraform-provider-tfe/pull/484))
* d/workspace_ids: Add `exclude_tags` to `tfe_workspace_ids` attributes ([#523](https://github.com/hashicorp/terraform-provider-tfe/pull/523))

## 0.31.0 (April 21, 2022)

BUG FIXES:
* Sensitive values within certain Authorization headers are now redacted from TRACE and DEBUG logs ([#479](https://github.com/hashicorp/terraform-provider-tfe/pull/479))
* r/tfe_variable_set: Clarified and fixed variable_set documentation and examples ([#473](https://github.com/hashicorp/terraform-provider-tfe/pull/473)) and ([#472](https://github.com/hashicorp/terraform-provider-tfe/pull/472))

FEATURES:
* r/team, d/team: Add sso_team_id to the tfe_team attributes ([#457](https://github.com/hashicorp/terraform-provider-tfe/pull/457))
* **New Authentication Method**: Host-specific TF_TOKEN_... variable can be used for token authentication. See provider documentation for details. ([#477](https://github.com/hashicorp/terraform-provider-tfe/pull/477))

## 0.30.2 (April 01, 2022)

BUG FIXES:
* r/tfe_variable_set: Fixed import documentation and examples ([#466](https://github.com/hashicorp/terraform-provider-tfe/pull/466))
* r/tfe_variable: Fixed import documentation and examples ([#466](https://github.com/hashicorp/terraform-provider-tfe/pull/466))

## 0.30.1 (April 01, 2022)

BUG FIXES:
* d/tfe_variable_set: Renamed variable_sets data source to variable_set in documentation ([#458](https://github.com/hashicorp/terraform-provider-tfe/pull/458))
* r/tfe_variable_set: Fixed examples in documentation for specifying workspace_ids ([#461](https://github.com/hashicorp/terraform-provider-tfe/pull/461))
* r/tfe_variable_set: Fixed examples in documentation for variable_set_id ([#462](https://github.com/hashicorp/terraform-provider-tfe/pull/462))

## 0.30.0 (March 29, 2022)

FEATURES:
* **New Resource**: `tfe_variable` ([#452](https://github.com/hashicorp/terraform-provider-tfe/pull/452))
* **New Resource**: `tfe_variable_set` ([#452](https://github.com/hashicorp/terraform-provider-tfe/pull/452))
* **New Data Sources**: d/tfe_variable_set, d/tfe_variables ([#452](https://github.com/hashicorp/terraform-provider-tfe/pull/452))

## 0.29.0 (March 24, 2022)

BUG FIXES:
* r/ssh_key: Removed ability to update ssh value, which never worked ([#432](https://github.com/hashicorp/terraform-provider-tfe/pull/432))

ENHANCEMENTS:
* r/team: Add `manage_providers` and `manage_modules` attributes to resource schema ([#431](https://github.com/hashicorp/terraform-provider-tfe/pull/431))
* Update go-tfe dependency to version 1.0.0 ([#450](https://github.com/hashicorp/terraform-provider-tfe/pull/450))

## 0.28.1 (February 04, 2022)

BUG FIXES:
* d/terraform_version: Backwards compatibility fix for importing Terraform versions from TFE installations that don't support filtering
  Terraform versions ([#427](https://github.com/hashicorp/terraform-provider-tfe/pull/427))

## 0.28.0 (February 02, 2022)

FEATURES:
* **New Resource**: `tfe_terraform_version` ([#400](https://github.com/hashicorp/terraform-provider-tfe/pull/400))
* **New Resource**: `tfe_organization_module_sharing` ([#425](https://github.com/hashicorp/terraform-provider-tfe/pull/425))

ENHANCEMENTS:
* r/workspace: Add support for importing workspaces using <ORGANIZATION NAME>/<WORKSPACE NAME> pair ([#401](https://github.com/hashicorp/terraform-provider-tfe/pull/401))
* r/team: Show entitlement error when creating teams ([#418](https://github.com/hashicorp/terraform-provider-tfe/pull/418))
* Bump `go-tfe` dependency to `0.24.0`

BUG FIXES:
* d/workspace_ids: Fix plugin crash when providing empty strings to `names` argument ([#421](https://github.com/hashicorp/terraform-provider-tfe/pull/421))
* r/workspace: Fix `trigger_prefixes` and `remote_state_consumer_ids` were appearing as workspace drift after being defaulted by the API to empty lists ([#423](https://github.com/hashicorp/terraform-provider-tfe/pull/423))

## 0.27.1 (January 25, 2022)

BUG FIXES:
* d/workspace: Fixed an issue with remote state consumers were being populated with all workspaces when
  global_remote_state is true. When global_remote_state is true, it's safe to assume that all workspace
  state can be read ([#414](https://github.com/hashicorp/terraform-provider-tfe/pull/414))

## 0.27.0 (December 15, 2021)

FEATURES:
* **New Data Source:** d/tfe_variables ([#369](https://github.com/hashicorp/terraform-provider-tfe/pull/369))

ENHANCEMENTS:
* r/organization: Added
  `send_passing_statuses_for_untriggered_speculative_plans`, which can be useful if large numbers of
  untriggered workspaces are exhausting request limits for connected version control service
  providers like GitHub. ([#386](https://github.com/hashicorp/terraform-provider-tfe/pull/386))
* r/oauth_client: Added `key`, `secret`, and `rsa_public_key` arguments, used for configuring
  BitBucket Server and Azure DevOps Server. ([#395](https://github.com/hashicorp/terraform-provider-tfe/pull/395))
* Improved discovery and loading of credentials from Terraform configuration files; the provider
  will attempt to use Terraform CLI's authentication with Terraform Cloud/Enterprise for its own
  authentication, when present. ([#360](https://github.com/hashicorp/terraform-provider-tfe/pull/360))

BUG FIXES:
* r/workspace: Fixed an issue with remote state consumer relationships on workspaces where the provider would not
  follow pagination and only the first 20 results would be read correctly. ([#367](https://github.com/hashicorp/terraform-provider-tfe/pull/367))
* r/tfe_variable: Fixed an issue where updating sensitive attributes would just surface the
  underlying correct error (they must be recreated) instead of allowing Terraform to intelligently
  replace the resource as part of its execution plan. ([#394](https://github.com/hashicorp/terraform-provider-tfe/pull/394))

## 0.26.1 (September 04, 2021)

BUG FIXES:
* Fixed a regression introduced in 0.26.0 where explicitly specifying a hostname became erroneously required, when it should
  default to app.terraform.io (Terraform Cloud) ([#354](https://github.com/hashicorp/terraform-provider-tfe/pull/354))
* d/workspace_ids: Fixed issue with `names` and `tag_names` not validating correctly ([#358](https://github.com/hashicorp/terraform-provider-tfe/pull/358))

## 0.26.0 (September 02, 2021)

FEATURES:
* **New Data Sources:** d/tfe_organizations, d/tfe_organization [#320](https://github.com/hashicorp/terraform-provider-tfe/pull/320).
* Add support for enabling structured run outputs in a `tfe_workspace` [#330](https://github.com/hashicorp/terraform-provider-tfe/pull/330).
* **New Data Source**: Introduces `tfe_slug` used to represent configuration files.
  on local file system [#333](https://github.com/hashicorp/terraform-provider-tfe/pull/333).
* Add functionality in `tfe_policy_set` to allow uploading of local policies [#333](https://github.com/hashicorp/terraform-provider-tfe/pull/333).
* **New Data Source**: Introduces `tfe_outputs` to retrieve state outputs for a Workspace.
* r/workspace: Added `tag_names` argument to set tags for a Workspace.
* d/workspace: Added `tag_names` to the data returned for a Workspace.
* d/workspace_ids: Added `tag_names` as a search option to find Workspaces by tag name.

ENHANCEMENTS:
* Use Golang 1.17 [#341](https://github.com/hashicorp/terraform-provider-tfe/pull/341).

## 0.25.3 (May 18, 2021)

BUG FIXES:
* d/ip_ranges: Fixes an issue in the upstream client where accessing this datasource would
  erroneously change the state of the client and cause subsequent requests in plans to fail with
  incorrect URLs. [#316](https://github.com/hashicorp/terraform-provider-tfe/pull/316)

## 0.25.2 (May 06, 2021)

BUG FIXES:
d/tfe_workspace: Fix remote state consumer regression for Terraform Enterprise ([#311](https://github.com/hashicorp/terraform-provider-tfe/pull/311))

NOTES:
* This release includes an additional fix for the regression introduced in v0.25.0
  to address errors for anyone using the `tfe_workspace` data source with a Terraform
  Enterprise version earlier than v20210401-1.


## 0.25.1 (April 30, 2021)

BUG FIXES:
* r/workspace: Fix remote state consumer regression for Terraform Enterprise ([#303](https://github.com/hashicorp/terraform-provider-tfe/pull/303))
* r/organization: Ignore diffs in name case sensitivity ([#300](https://github.com/hashicorp/terraform-provider-tfe/pull/300))

NOTES:
* This release includes a fix for a major regression from a backwards incompatible change
  erroneously introduced in v0.25.0, where any Terraform Enterprise version < v20210401-1 would
  experience failures using the tfe_workspace resource.

## 0.25.0 (April 29, 2021)

BREAKING CHANGES:
* d/tfe_workspace: Removed deprecated `external_id` attribute. Use `id` instead ([#295](https://github.com/hashicorp/terraform-provider-tfe/pull/295))
* d/tfe_workspace_ids: Removed deprecated `external_ids` attribute. Use `ids` instead ([#295](https://github.com/hashicorp/terraform-provider-tfe/pull/295))
* r/tfe_workspace: Removed deprecated `external_id` attribute. Use `id` instead ([#295](https://github.com/hashicorp/terraform-provider-tfe/pull/295))

ENHANCEMENTS:
* Use Go 1.16 to provide support for Apple Silicon (darwin/arm64) ([#288](https://github.com/hashicorp/terraform-provider-tfe/pull/288))
* Add Manage Policy Overrides permission for teams ([#285](https://github.com/hashicorp/terraform-provider-tfe/pull/285))
* r/tfe_workspace: Add remote state consumer functionality ([#292](https://github.com/hashicorp/terraform-provider-tfe/pull/292))
* r/tfe_workspace: Added description parameter to TFE workspace ([#271](https://github.com/hashicorp/terraform-provider-tfe/pull/271))
* d/tfe_workspace: Added new workspace fields from the API ([#287](https://github.com/hashicorp/terraform-provider-tfe/pull/287))
* d/tfe_workspace: Added `branch` attribute to `vcs_repo` block ([#290](https://github.com/hashicorp/terraform-provider-tfe/pull/290))
* Improved error message for missing token ([#273](https://github.com/hashicorp/terraform-provider-tfe/pull/273))

NOTES:
* You will need to migrate to the new attributes in your configuration to update to the latest
  version of this provider. The tfe_workspace resource will continue to migrate old workspace
  resources in state (schema version 0, using `external_id`) to new ones (schema version 1, using `id`) for
  the foreseeable future and will only be removed in a breaking major version (likely v1.0.0). More information
  about these deprecations can be found in the description of [#295](https://github.com/hashicorp/terraform-provider-tfe/pull/295)

## 0.24.0 (January 22, 2021)

BREAKING CHANGES:
* Support for Terraform version 0.11 and prior has ended. Terraform version 0.12+ is required. This is a result of
  updating the provider to use version 2.0 of the [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) ([#246](https://github.com/hashicorp/terraform-provider-tfe/pull/246))
* d/tfe_workspace_ids: Changed `ids` attribute to return immutable workspace IDs (`ws-<RANDOM STRING>`) ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))
* r/tfe_notification_configuration: Removed deprecated `workspace_external_id` attribute, preferring `workspace_id` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))
* r/tfe_policy_set: Removed deprecated `workspace_external_ids` attribute, preferring `workspace_ids` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))
* r/tfe_run_trigger: Removed deprecated `workspace_external_id` attribute, preferring `workspace_id` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))

FEATURES:
* **New Resource:** r/tfe_agent_token ([#259](https://github.com/hashicorp/terraform-provider-tfe/pull/259))
* **New Data Source:** d/tfe_ip_ranges ([#262](https://github.com/hashicorp/terraform-provider-tfe/pull/262))

ENHANCEMENTS:
* d/tfe_workspace: Added deprecation warning to the `external_id` attribute, preferring `id` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))
* d/tfe_workspace_ids: Added deprecation warning to the `external_ids` attribute, preferring `ids` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))
* r/tfe_workspace: Added deprecation warning to the `external_id` attribute, preferring `id` instead ([#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253))

NOTES:
* All deprecated attributes will be removed 3 months after the release of v0.24.0 (April 21, 2021). After this
  deprecation period, you will need to migrate to the preferred attributes to update to the latest version of this
  provider. More information about these deprecations can be found in the description of [#253](https://github.com/hashicorp/terraform-provider-tfe/pull/253)
* d/tfe_workspace: The deprecation warning for the `external_id` attribute will not go away until the attribute is
  removed in a future version.  This is due to a [limitation of the Terraform
  SDK](https://github.com/hashicorp/terraform/issues/7569) for deprecation warnings on attributes that aren't specified
  in a configuration. If you have already changed all references to this data source's `external_id` attribute to the
  `ids` attribute, you can ignore the warning.
* d/tfe_workspace_ids: The deprecation warning for the `external_ids` attribute will not go away until the attribute is
  removed in a future version.  This is due to a [limitation of the Terraform
  SDK](https://github.com/hashicorp/terraform/issues/7569) for deprecation warnings on attributes that aren't specified
  in a configuration. If you have already changed all references to this data source's `external_ids` attribute to the
  `ids` attribute, you can ignore the warning.


## 0.23.0 (November 20, 2020)

FEATURES:
* **New Resource:** r/tfe_agent_pool ([#242](https://github.com/hashicorp/terraform-provider-tfe/pull/242)) Includes
  the ability to import existing agent pools via ID.
* **New Data Source:** d/tfe_agent_pool ([#242](https://github.com/hashicorp/terraform-provider-tfe/pull/242))

ENHANCEMENTS:
* r/tfe_workspace: Added `execution_mode` argument, succeeding the existing `operations` boolean (which is now
  deprecated) ([#242](https://github.com/hashicorp/terraform-provider-tfe/pull/242)) This new argument, along with
  `agent_pool_id`, allows for configuring workspaces to use Terraform Cloud Agents
  (https://www.terraform.io/docs/cloud/agents).
* r/tfe_workspace: Added `allow_destroy_plan`, which determines if destroy plans can be queued on the workspace ([#245](https://github.com/hashicorp/terraform-provider-tfe/pull/245))
* r/tfe_organization: Added `cost_estimation_enabled`, which determines if the cost estimation feature is enabled for all workspaces in the organization. ([#239](https://github.com/hashicorp/terraform-provider-tfe/pull/239))
* Added provider configuration option `ssl_skip_verify`, to allow users to skip certificate verifications if their
  environment is appropriate for it (note that in general, this is not recommended and the default value of `true`
  should be used). ([#95](https://github.com/hashicorp/terraform-provider-tfe/pull/95))

BUG FIXES:
* r/tfe_team_access: Fixed an erroneous error message seen when a workspace could not be retrieved from the API ([#233](https://github.com/hashicorp/terraform-provider-tfe/pull/233))

NOTES:
  * Go 1.14 is now being used for development, along with Go modules.
  * Several documentation improvements have been made in this release.

## 0.22.0 (October 07, 2020)

FEATURES:
* **New Data Source:** d/tfe_oauth_client ([#212](https://github.com/hashicorp/terraform-provider-tfe/pull/212))

ENHANCEMENTS:
* r/tfe_variable: Changes to the key of a sensitive variable will result in the deletion of the old variable and the creation of a new one ([#175](https://github.com/hashicorp/terraform-provider-tfe/pull/175))
* r/tfe_workspace: Adds support for the speculative_enabled argument to tfe_workspace ([#210](https://github.com/hashicorp/terraform-provider-tfe/pull/210))

BUG FIXES:
* r/tfe_registry_module: Prevent a possible race condition when creating modules in the registry. ([#215](https://github.com/hashicorp/terraform-provider-tfe/pull/215))
* r/tfe_run_trigger: Retry when a "locked" error is returned ([#178](https://github.com/hashicorp/terraform-provider-tfe/pull/178))
* r/tfe_workspace: Fixed a logic bug that prevented non-default branch names to be imported. ([#220](https://github.com/hashicorp/terraform-provider-tfe/pull/220))
* r/tfe_workspace: Prevent the provider from crashing when encountering empty trigger prefixes. ([#223](https://github.com/hashicorp/terraform-provider-tfe/pull/223))
* r/tfe_workspace_variable: Remove the variable from the state if the workspace containing it has been deleted via the UI. ([#227](https://github.com/hashicorp/terraform-provider-tfe/pull/227))

## 0.21.0 (August 19, 2020)

ENHANCEMENTS:
* r/tfe_policy_set: Added a validation for the `name` attribute so that invalid policy set names are caught at plan time ([#168](https://github.com/hashicorp/terraform-provider-tfe/pull/168))

NOTES:
* This validation matches the requirements specified by the [Terraform Cloud API](https://www.terraform.io/docs/cloud/api/policy-sets.html#request-body). Policy set names can only include letters, numbers, -, and \_.

## 0.20.0 (July 17, 2020)

FEATURES:
* **New Resource:** r/tfe_registry_module ([#191](https://github.com/hashicorp/terraform-provider-tfe/pull/191))
* **New Data Source:** d/tfe_organization_membership ([#191](https://github.com/hashicorp/terraform-provider-tfe/pull/191))

ENHANCEMENTS:
* r/tfe_notification_configuration: Added support for email notification configuration by adding support for `destination_type` of `email` and associated schema attributes `email_user_ids` and (TFE only) `email_addresses` ([#191](https://github.com/hashicorp/terraform-provider-tfe/pull/191))
* r/tfe_organization_membership: Added ability to import organization memberships and added new computed attribute `user_id` ([#191](https://github.com/hashicorp/terraform-provider-tfe/pull/191))

NOTES:
* Using `destination_type` of `email` with resource `tfe_notification_configuration` requires using the provider with Terraform Cloud or an instance of Terraform Enterprise at least as recent as v202005-1.

## 0.19.0 (June 17, 2020)

FEATURES:
* r/tfe_team_access and d/tfe_team_access: Added support for custom workspace permissions ([#184](https://github.com/hashicorp/terraform-provider-tfe/pull/184))

BUG FIXES:
* r/tfe_policy_set: Fixes issue when updating Policy Set branch attribute ([#185](https://github.com/hashicorp/terraform-provider-tfe/pull/185))

## 0.18.1 (June 10, 2020)

ENHANCEMENTS:
* provider: Updated terraform-provider-sdk to 1.13.1 ([[#177](https://github.com/hashicorp/terraform-provider-tfe/pull/177)])

## 0.18.0 (June 03, 2020)

ENHANCEMENTS:
* d/tfe_workspace_ids: Added deprecation warning to the `ids` attribute, preferring `full_names` instead ([#182](https://github.com/hashicorp/terraform-provider-tfe/pull/182))
* r/tfe_notification_configuration: Added deprecation warning to the `workspace_external_id` attribute, preferring `workspace_id` instead ([#182](https://github.com/hashicorp/terraform-provider-tfe/pull/182))
* r/tfe_policy_set: Added deprecation warning to the `workspace_external_ids` attribute, preferring `workspace_ids` instead ([#182](https://github.com/hashicorp/terraform-provider-tfe/pull/182))
* r/tfe_run_trigger: Added deprecation warning to the `workspace_external_id` attribute, preferring `workspace_id` instead ([#182](https://github.com/hashicorp/terraform-provider-tfe/pull/182))

NOTES:
* All deprecated attributes will be removed 3 months after the release of v0.18.0. You will have until September 3, 2020 to migrate to the preferred attributes.
* More information about these deprecations can be found in the description of [#182](https://github.com/hashicorp/terraform-provider-tfe/pull/182)
* d/tfe_workspace_ids: The deprecation warning for the `ids` attribute will not go away until the attribute is removed in a future version.
This is due to a [limitation of the 1.0 version of the Terraform SDK](https://github.com/hashicorp/terraform/issues/7569) for deprecation warnings on attributes that aren't specified in a configuration.
If you have already changed all references to this data source's `ids` attribute to the new `full_names` attribute, you can ignore the warning.


## 0.17.1 (May 27, 2020)

BUG FIXES:
* r/tfe_team: Fixed a panic occurring with importing Owners teams on Free TFC organizations which do not include visible organization access. ([#181](https://github.com/hashicorp/terraform-provider-tfe/pull/181))

## 0.17.0 (May 21, 2020)

ENHANCEMENTS:
* r/tfe_team: Added support for organization-level permissions and visibility on teams. ([#155](https://github.com/hashicorp/terraform-provider-tfe/pull/155))

## 0.16.2 (May 12, 2020)

BUG FIXES:
* r/tfe_workspace: Allow VCS repo to be removed from a workspace when it has been removed from the configuration. ([#173](https://github.com/hashicorp/terraform-provider-tfe/pull/173))

## 0.16.1 (April 28, 2020)

BUG FIXES:
* r/tfe_workspace: Running a plan/apply when a workspace has been deleted outside of
  terraform no longer causes a panic. ([#162](https://github.com/hashicorp/terraform-provider-tfe/pull/162))

## 0.16.0 (April 14, 2020)

FEATURES:

- **New Resource**: `tfe_organization_membership` ([#154](https://github.com/hashicorp/terraform-provider-tfe/pull/154))
- **New Resource**: `tfe_team_organization_member` ([#154](https://github.com/hashicorp/terraform-provider-tfe/pull/154))

## 0.15.1 (March 25, 2020)

ENHANCEMENTS:
* r/tfe_workspace: Migrate ID from <organization>/<workspace> to opaque external_id ([#106](https://github.com/hashicorp/terraform-provider-tfe/pull/106))
* r/tfe_variable: Migrate workspace_id from <organization>/<workspace> to opaque external_id ([#106](https://github.com/hashicorp/terraform-provider-tfe/pull/106))
* r/tfe_team_access: Migrate workspace_id from <organization>/<workspace> to opaque external_id ([#106](https://github.com/hashicorp/terraform-provider-tfe/pull/106))

## 0.15.0 (March 25, 2020)

## 0.14.1 (March 04, 2020)

BUG FIXES:

* t/tfe_workspace: Issues with updating `working_directory` ([[#137](https://github.com/hashicorp/terraform-provider-tfe/pull/137)])
  and `trigger_prefixes` ([[#138](https://github.com/hashicorp/terraform-provider-tfe/pull/138)]) when removed from the configuration.
  Special note: if you have workspaces which are configured through the TFE provider, but have set the working directory or trigger prefixes manually, through the UI, you'll need to update your configuration.

## 0.14.0 (February 20, 2020)

FEATURES:

* **New Resource:** `tfe_run_trigger` ([[#132](https://github.com/hashicorp/terraform-provider-tfe/pull/132)])

## 0.13.0 (February 18, 2020)

ENHANCEMENTS:

* provider: Update to the standalone SDK ([[#130](https://github.com/hashicorp/terraform-provider-tfe/pull/130)])

## 0.12.1 (February 12, 2020)

BUG FIXES:

* provider: Lock the provider v2.2 for Terraform Enterprise ([[#127](https://github.com/hashicorp/terraform-provider-tfe/pull/127)])
This will warn users that this version of the provider does not support Terraform Enterprise versions < 202001-1

## 0.12.0 (February 11, 2020)

BREAKING CHANGES:

* r/tfe_variable: Update the workspace variable resource to utilize the "nested" routes that are now preferred ([[#123](https://github.com/hashicorp/terraform-provider-tfe/pull/123)])
This change is incompatible with Terraform Enterprise versions < 202001-1.

ENHANCEMENTS:

* **New Resource:** `tfe_policy_set_parameter` ([[#123](https://github.com/hashicorp/terraform-provider-tfe/pull/123)])
* r/tfe_variable: Add support for descriptions for workspace variables ([[#121](https://github.com/hashicorp/terraform-provider-tfe/pull/121)])

## 0.11.4 (December 13, 2019)

BUG FIXES:

r/tfe_oauth_client: Issue with using private_key and validation check ([[#113]](https://github.com/hashicorp/terraform-provider-tfe/pull/113))

## 0.11.3 (December 10, 2019)

ENHANCEMENTS:

* r/tfe_oauth_client: Adding support for Azure DevOps Server and Azure DevOps Services ([[#99](https://github.com/hashicorp/terraform-provider-tfe/pull/99)])

## 0.11.2 (December 10, 2019)

ENHANCEMENTS:

* provider: Retry requests which result in server errors ([[#109](https://github.com/hashicorp/terraform-provider-tfe/pull/109)])

## 0.11.1 (September 27, 2019)

ENHANCEMENTS:

* r/tfe_workspace: Adding support to configure execution mode ([[#92](https://github.com/hashicorp/terraform-provider-tfe/pull/92)])

## 0.11.0 (August 19, 2019)

FEATURES:

* **New Resource:** `tfe_notification_configuration` ([[#86](https://github.com/hashicorp/terraform-provider-tfe/pull/86)])

## 0.10.1 (June 26, 2019)

BUG FIXES:

* r/tfe_workspace: Ensure that file-triggers-enabled and trigger-prefixes fields are updated when changed ([#81](https://github.com/hashicorp/terraform-provider-tfe/pull/81))

## 0.10.0 (June 20, 2019)

ENHANCEMENTS:

* r/tfe_policy_set: Added support for VCS policy sets. ([#80](https://github.com/hashicorp/terraform-provider-tfe/issues/80))

## 0.9.1 (June 05, 2019)

ENHANCEMENTS:

* r/tfe_workspace: Add monorepo filtering workspace config fields ([#77](https://github.com/hashicorp/terraform-provider-tfe/pull/77))
* provider: Add support for TFE_HOSTNAME and TFE_TOKEN environment variables ([#78](https://github.com/hashicorp/terraform-provider-tfe/pull/78), fixes [#31](https://github.com/hashicorp/terraform-provider-tfe/issues/31))

## 0.9.0 (May 23, 2019)

IMPROVEMENTS:

* The provider is now compatible with Terraform v0.12, while retaining compatibility with prior versions.

## 0.8.2 (April 08, 2019)

BUG FIXES:

* d/tfe_workspace: Set the correct workspace ID ([#74](https://github.com/hashicorp/terraform-provider-tfe/issues/74))

## 0.8.1 (March 26, 2019)

BUG FIXES:

* provider: Update the vendor directory so it's in sync with the versions defined in `go.mod` ([#73](https://github.com/hashicorp/terraform-provider-tfe/issues/73))

## 0.8.0 (March 26, 2019)

BUG FIXES:

* r/tfe_variable: Mark `value` as optional (defaults to `""`) to match TFE API behavior ([#72](https://github.com/hashicorp/terraform-provider-tfe/issues/72))

## 0.7.1 (February 15, 2019)

BUG FIXES:

* r/tfe_workspace: Add a check when migrating `vcs_repo` from a set to a list ([#64](https://github.com/hashicorp/terraform-provider-tfe/issues/64))

## 0.7.0 (February 14, 2019)

ENHANCEMENTS:

* provider: Enable request/response logging ([#55](https://github.com/hashicorp/terraform-provider-tfe/issues/55))
* provider: Report detailed service discovery and version constraints information ([#61](https://github.com/hashicorp/terraform-provider-tfe/issues/61))
* r/tfe_workspace: Try to find a workspace by external ID before removing it ([#51](https://github.com/hashicorp/terraform-provider-tfe/issues/51))
* r/tfe_workspace: Use a list instead of a set for a workspace `vcs_repo` ([#53](https://github.com/hashicorp/terraform-provider-tfe/issues/53))

## 0.6.0 (January 08, 2019)

FEATURES:

* **New resource**: `tfe_oauth_client` ([#42](https://github.com/hashicorp/terraform-provider-tfe/issues/42))
* **New data source**: `tfe_ssh_key` ([#43](https://github.com/hashicorp/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_team` ([#43](https://github.com/hashicorp/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_team_access` ([#43](https://github.com/hashicorp/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_workspace` ([#43](https://github.com/hashicorp/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_workspace_ids` ([#43](https://github.com/hashicorp/terraform-provider-tfe/issues/43))

## 0.5.0 (December 12, 2018)

ENHANCEMENTS:

* r/tfe_workspace: Support queuing all runs for new workspaces ([#41](https://github.com/hashicorp/terraform-provider-tfe/issues/41))

## 0.4.0 (November 27, 2018)

ENHANCEMENTS:

* r/tfe_workspace: Support assigning an SSH key to a workspace ([#38](https://github.com/hashicorp/terraform-provider-tfe/issues/38))

## 0.3.0 (November 13, 2018)

FEATURES:

* **New resource**: `tfe_policy_set` ([#33](https://github.com/hashicorp/terraform-provider-tfe/issues/33))

ENHANCEMENTS:

* `go-tfe` now includes logic to throttle requests preventing rate limit errors ([#34](https://github.com/hashicorp/terraform-provider-tfe/issues/34))

BUG FIXES:

* r/tfe_workspace: Fix a bug that prevented to set `auto-apply` to false ([#30](https://github.com/hashicorp/terraform-provider-tfe/issues/30))

## 0.2.0 (September 20, 2018)

NOTES:

* r/tfe_workspace: The format of the internal ID used to track workspaces
  is changed to be more inline with other representations of the same ID. The ID
  should be converted automatically during an `apply`, but the conversion can also
  be triggered manually by running `terraform refresh` when it causes issues.

FEATURES:

* Add `terraform import` support to all (except `tfe_ssh_key`) resources ([#20](https://github.com/hashicorp/terraform-provider-tfe/issues/20))

ENHANCEMENTS:

* r/tfe_workspace: Export the Terraform Enterprise workspace ID ([#21](https://github.com/hashicorp/terraform-provider-tfe/issues/21))

## 0.1.0 (August 14, 2018)

Initial release.
