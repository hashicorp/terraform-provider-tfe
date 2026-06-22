resource "tfe_tag_policy_set" "env_any" {
  policy_set_id = tfe_policy_set.test.id
  key           = "env"
}
