variable "session_token" {
  type      = string
  ephemeral = true
}

resource "tfe_test_variable" "tf_test_test_variable" {
  key              = "key_test"
  value_wo         = var.session_token
  value_wo_version = 1
  description      = "some description"
  category         = "env"
  organization     = tfe_organization.test_org.name
  module_name      = tfe_registry_module.test_module.name
  module_provider  = tfe_registry_module.test_module.module_provider
}
