# For key and value tags, via <POLICY SET ID>/<TAG KEY>/<TAG VALUE>
terraform import tfe_tag_policy_set_exclusion.test 'polset-abc123/env/staging'

# For key-only tags, via <POLICY SET ID>/<TAG KEY>
terraform import tfe_tag_policy_set_exclusion.test 'polset-abc123/key-only'