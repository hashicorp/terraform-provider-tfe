# In the example below, assume we have outputs defined in a my-org/my-workspace

ephemeral "tfe_outputs" "foo" {
  organization = "my-org"
  workspace    = "my-workspace"
}

provider "vault" {
  auth_login {
    path = "auth/approle/login"

    parameters = {
      role_id   = ephemeral_tfe_outputs.foo.values.vault_role_id
      secret_id = ephemeral_tfe_outputs.foo.values.vault_secret_id
    }
  }
}
