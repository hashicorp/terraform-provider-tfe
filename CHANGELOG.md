## 0.2.0 (Unreleased)

NOTES:

* resource/tfe_workspace: The format of the internal ID used to track workspaces
  is changed to be more inline with other representations of the same ID. The ID
  should be converted automatically during an `apply`, but the conversion can also
  be triggered manually by running `terraform refresh` when it causes a issues.

FEATURES:

* Add `terraform import` support to all (except `tfe_ssh_key`) resources [GH-20]

ENHANCEMENTS:

* resource/tfe_workspace: Export the Terraform Enterprise workspace ID [GH-21]

## 0.1.0 (August 14, 2018)

Initial release.
