# via <ORGANIZATION>/<REGISTRY NAME>/<NAMESPACE>/<PROVIDER NAME>/<VERSION>
# For a private provider version:
terraform import tfe_registry_provider_version.example my-org-name/private/my-org-name/my-provider/1.0.0

# For a public provider version:
terraform import tfe_registry_provider_version.example my-org-name/public/hashicorp/aws/5.0.0
