# Basic usage

resource "tfe_vault_oidc_configuration" "example" {
  address        = "https://my-vault-cluster-public-vault-659decf3.b8298d98.z1.hashicorp.cloud:8200"
  role_name      = "vault-role-name"
  namespace      = "admin"
  auth_path      = "jwt-auth-path"
  encoded_cacert = ""
  organization   = "my-org-name"
}
