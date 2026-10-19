# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# For key and value tags, via <POLICY SET ID>/<TAG KEY>/<TAG VALUE>
terraform import tfe_tag_policy_set_exclusion.test 'polset-abc123/env/staging'

# For key-only tags, via <POLICY SET ID>/<TAG KEY>
terraform import tfe_tag_policy_set_exclusion.test 'polset-abc123/key-only'