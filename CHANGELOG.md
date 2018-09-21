## 0.2.1 (Unreleased)
## 0.2.0 (September 20, 2018)

NOTES:

* resource/tfe_workspace: The format of the internal ID used to track workspaces
  is changed to be more inline with other representations of the same ID. The ID
  should be converted automatically during an `apply`, but the conversion can also
  be triggered manually by running `terraform refresh` when it causes issues.

FEATURES:

* Add `terraform import` support to all (except `tfe_ssh_key`) resources ([#20](https://github.com/terraform-providers/terraform-provider-tfe/issues/20))

ENHANCEMENTS:

* resource/tfe_workspace: Export the Terraform Enterprise workspace ID ([#21](https://github.com/terraform-providers/terraform-provider-tfe/issues/21))

## 0.1.0 (August 14, 2018)

Initial release.
