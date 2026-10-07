# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# via <AGENT POOL ID>
terraform import tfe_agent_pool.test apool-rW0KoLSlnuNb5adB

# via <ORGANIZATION NAME>/<AGENT POOL NAME>
terraform import tfe_agent_pool.test my-org-name/my-agent-pool-name