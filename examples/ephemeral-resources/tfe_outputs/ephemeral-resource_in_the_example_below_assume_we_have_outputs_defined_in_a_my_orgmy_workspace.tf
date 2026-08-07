# Using the `tfe_outputs` ephemeral resource, the outputs `vault_role_id` and `vault_secret_id` can be used to configure a vault provider instance as seen below:
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
