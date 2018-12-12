## 0.5.0 (Unreleased)

ENHANCEMENTS:

* r/tfe_workspace: Support queuing all runs for new workspaces [GH-41]

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
