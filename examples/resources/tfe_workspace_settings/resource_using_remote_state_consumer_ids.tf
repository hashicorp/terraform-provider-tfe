# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

resource "tfe_workspace" "test" {
  for_each = toset(["qa", "production"])
  name     = "${each.value}-test"
}

resource "tfe_workspace_settings" "test-settings" {
  for_each                  = toset(["qa", "production"])
  workspace_id              = tfe_workspace.test[each.value].id
  global_remote_state       = false
  project_remote_state      = false
  remote_state_consumer_ids = toset(compact([each.value == "production" ? tfe_workspace.test["qa"].id : ""]))
}
