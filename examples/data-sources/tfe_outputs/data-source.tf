# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# In the example below, assume we have outputs defined in a my-org/my-workspace

data "tfe_outputs" "foo" {
  organization = "my-org"
  workspace    = "my-workspace"
}

resource "random_id" "vpc_id" {
  keepers = {
    # Generate a new ID any time the value of 'bar' in workspace 'my-org/my-workspace' changes.
    bar = data.tfe_outputs.foo.values.bar
  }

  byte_length = 8
}
