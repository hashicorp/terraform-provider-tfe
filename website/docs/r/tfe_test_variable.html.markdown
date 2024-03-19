---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_test_variable"
description: |-
  Manages environmet variables used for testing by modules in the Private Module Registry.
---

# tfe_test_variable

Creates, updates and destroys environment variables used for testing in the Private Module Registry.

## Example Usage

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}
  
resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "my-token-123"
  service_provider = "github"
}

resource "tfe_registry_module" "test" {
  organization     = "test-module"
  vcs_repo {
  display_identifier = "GH_NAME/REPO_NAME"
  identifier         = "GH_NAME/REPO_NAME"
  oauth_token_id     = tfe_oauth_client.test.oauth_token_id
  branch             = "main"
  tags				 = false
}
  test_config {
	tests_enabled = true
  }
}

resource "tfe_test_variable" "test" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  organization = tfe_organization.test.name
  module_name = tfe_registry_module.test.name
  module_provider = tfe_registry_module.test.module_provider
}
```
