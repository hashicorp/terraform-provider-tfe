variable "session_token" {
  description = "The session token for use in the write only field"
}

resource "tfe_organization" "test_org" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_oauth_client" "test_client" {
  organization     = tfe_organization.test_org.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-token-123"
  service_provider = "github"
}

resource "tfe_registry_module" "test_module" {
  organization = "test-module"
  vcs_repo {
    display_identifier = "GH_NAME/REPO_NAME"
    identifier         = "GH_NAME/REPO_NAME"
    oauth_token_id     = tfe_oauth_client.test_client.oauth_token_id
    branch             = "main"
    tags               = false
  }
  test_config {
    tests_enabled = true
  }
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
