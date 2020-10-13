## 0.23.0 (Unreleased)
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
* This validation matches the requirements specified by the [Terraform Cloud API](https://www.terraform.io/docs/cloud/api/policy-sets.html#request-body). Policy set names can only include letters, numbers, -, and _.

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
