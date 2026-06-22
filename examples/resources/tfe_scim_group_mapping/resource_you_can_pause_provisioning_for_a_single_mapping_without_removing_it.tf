resource "tfe_scim_group_mapping" "engineering" {
  team_id       = tfe_team.engineering.id
  scim_group_id = data.tfe_scim_group.engineering.id
  paused        = true
}
