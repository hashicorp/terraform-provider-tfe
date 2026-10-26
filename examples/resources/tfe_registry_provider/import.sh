# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

# via <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>
# For a private provider:
terraform import tfe_registry_provider.example my-org-name/private/my-org-name/my-provider

# For a public provider:
terraform import tfe_registry_provider.example my-org-name/public/hashicorp/aws