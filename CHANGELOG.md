## 0.14.0 (Unreleased)

FEATURES:

* **New Resource:** `tfe_run_trigger` ([[#132](https://github.com/terraform-providers/terraform-provider-tfe/pull/132)])

## 0.13.0 (February 18, 2020)

ENHANCEMENTS:

* provider: Update to the standalone SDK ([[#130](https://github.com/terraform-providers/terraform-provider-tfe/pull/130)])

## 0.12.1 (February 12, 2020)

BUG FIXES:

* provider: Lock the provider v2.2 for Terraform Enterprise ([[#127](https://github.com/terraform-providers/terraform-provider-tfe/pull/127)])
This will warn users that this version of the provider does not support Terraform Enterprise versions < 202001-1

## 0.12.0 (February 11, 2020)

BREAKING CHANGES:

* r/tfe_variable: Update the workspace variable resource to utilize the "nested" routes that are now preferred ([[#123](https://github.com/terraform-providers/terraform-provider-tfe/pull/123)])
This change is incompatible with Terraform Enterprise versions < 202001-1. 

ENHANCEMENTS:

* **New Resource:** `tfe_policy_set_parameter` ([[#123](https://github.com/terraform-providers/terraform-provider-tfe/pull/123)])
* r/tfe_variable: Add support for descriptions for workspace variables ([[#121](https://github.com/terraform-providers/terraform-provider-tfe/pull/121)])

## 0.11.4 (December 13, 2019)

BUG FIXES:

r/tfe_oauth_client: Issue with using private_key and validation check ([[#113]](https://github.com/terraform-providers/terraform-provider-tfe/pull/113))

## 0.11.3 (December 10, 2019)

ENHANCEMENTS:

* r/tfe_oauth_client: Adding support for Azure DevOps Server and Azure DevOps Services ([[#99](https://github.com/terraform-providers/terraform-provider-tfe/pull/99)])

## 0.11.2 (December 10, 2019)

ENHANCEMENTS:

* provider: Retry requests which result in server errors ([[#109](https://github.com/terraform-providers/terraform-provider-tfe/pull/109)])

## 0.11.1 (September 27, 2019)

ENHANCEMENTS:

* r/tfe_workspace: Adding support to configure execution mode ([[#92](https://github.com/terraform-providers/terraform-provider-tfe/pull/92)])

## 0.11.0 (August 19, 2019)

FEATURES:

* **New Resource:** `tfe_notification_configuration` ([[#86](https://github.com/terraform-providers/terraform-provider-tfe/pull/86)])

## 0.10.1 (June 26, 2019)

BUG FIXES:

* r/tfe_workspace: Ensure that file-triggers-enabled and trigger-prefixes fields are updated when changed ([#81](https://github.com/terraform-providers/terraform-provider-tfe/pull/81))

## 0.10.0 (June 20, 2019)

ENHANCEMENTS:

* r/tfe_policy_set: Added support for VCS policy sets. ([#80](https://github.com/terraform-providers/terraform-provider-tfe/issues/80))

## 0.9.1 (June 05, 2019)

ENHANCEMENTS:

* r/tfe_workspace: Add monorepo filtering workspace config fields ([#77](https://github.com/terraform-providers/terraform-provider-tfe/pull/77))
* provider: Add support for TFE_HOSTNAME and TFE_TOKEN environment variables ([#78](https://github.com/terraform-providers/terraform-provider-tfe/pull/78), fixes [#31](https://github.com/terraform-providers/terraform-provider-tfe/issues/31))

## 0.9.0 (May 23, 2019)

IMPROVEMENTS:

* The provider is now compatible with Terraform v0.12, while retaining compatibility with prior versions.

## 0.8.2 (April 08, 2019)

BUG FIXES:

* d/tfe_workspace: Set the correct workspace ID ([#74](https://github.com/terraform-providers/terraform-provider-tfe/issues/74))

## 0.8.1 (March 26, 2019)

BUG FIXES:

* provider: Update the vendor directory so it's in sync with the versions defined in `go.mod` ([#73](https://github.com/terraform-providers/terraform-provider-tfe/issues/73))

## 0.8.0 (March 26, 2019)

BUG FIXES:

* r/tfe_variable: Mark `value` as optional (defaults to `""`) to match TFE API behavior ([#72](https://github.com/terraform-providers/terraform-provider-tfe/issues/72))

## 0.7.1 (February 15, 2019)

BUG FIXES:

* r/tfe_workspace: Add a check when migrating `vcs_repo` from a set to a list ([#64](https://github.com/terraform-providers/terraform-provider-tfe/issues/64))

## 0.7.0 (February 14, 2019)

ENHANCEMENTS:

* provider: Enable request/response logging ([#55](https://github.com/terraform-providers/terraform-provider-tfe/issues/55))
* provider: Report detailed service discovery and version constraints information ([#61](https://github.com/terraform-providers/terraform-provider-tfe/issues/61))
* r/tfe_workspace: Try to find a workspace by external ID before removing it ([#51](https://github.com/terraform-providers/terraform-provider-tfe/issues/51))
* r/tfe_workspace: Use a list instead of a set for a workspace `vcs_repo` ([#53](https://github.com/terraform-providers/terraform-provider-tfe/issues/53))

## 0.6.0 (January 08, 2019)

FEATURES:

* **New resource**: `tfe_oauth_client` ([#42](https://github.com/terraform-providers/terraform-provider-tfe/issues/42))
* **New data source**: `tfe_ssh_key` ([#43](https://github.com/terraform-providers/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_team` ([#43](https://github.com/terraform-providers/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_team_access` ([#43](https://github.com/terraform-providers/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_workspace` ([#43](https://github.com/terraform-providers/terraform-provider-tfe/issues/43))
* **New data source**: `tfe_workspace_ids` ([#43](https://github.com/terraform-providers/terraform-provider-tfe/issues/43))

## 0.5.0 (December 12, 2018)

ENHANCEMENTS:

* r/tfe_workspace: Support queuing all runs for new workspaces ([#41](https://github.com/terraform-providers/terraform-provider-tfe/issues/41))

## 0.4.0 (November 27, 2018)

ENHANCEMENTS:

* r/tfe_workspace: Support assigning an SSH key to a workspace ([#38](https://github.com/terraform-providers/terraform-provider-tfe/issues/38))

## 0.3.0 (November 13, 2018)

FEATURES:

* **New resource**: `tfe_policy_set` ([#33](https://github.com/terraform-providers/terraform-provider-tfe/issues/33))

ENHANCEMENTS:

* `go-tfe` now includes logic to throttle requests preventing rate limit errors ([#34](https://github.com/terraform-providers/terraform-provider-tfe/issues/34))

BUG FIXES:

* r/tfe_workspace: Fix a bug that prevented to set `auto-apply` to false ([#30](https://github.com/terraform-providers/terraform-provider-tfe/issues/30))

## 0.2.0 (September 20, 2018)

NOTES:

* r/tfe_workspace: The format of the internal ID used to track workspaces
  is changed to be more inline with other representations of the same ID. The ID
  should be converted automatically during an `apply`, but the conversion can also
  be triggered manually by running `terraform refresh` when it causes issues.

FEATURES:

* Add `terraform import` support to all (except `tfe_ssh_key`) resources ([#20](https://github.com/terraform-providers/terraform-provider-tfe/issues/20))

ENHANCEMENTS:

* r/tfe_workspace: Export the Terraform Enterprise workspace ID ([#21](https://github.com/terraform-providers/terraform-provider-tfe/issues/21))

## 0.1.0 (August 14, 2018)

Initial release.
