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
  organization     = "test-module"
  vcs_repo {
  display_identifier = "GH_NAME/REPO_NAME"
  identifier         = "GH_NAME/REPO_NAME"
  oauth_token_id     = tfe_oauth_client.test_client.oauth_token_id
  branch             = "main"
  tags				 = false
}
  test_config {
	tests_enabled = true
  }
}

resource "tfe_test_variable" "tf_test_test_variable" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  organization = tfe_organization.test_org.name
  module_name = tfe_registry_module.test_module.name
  module_provider = tfe_registry_module.test_module.module_provider
}
```

-> **Note:** Write-Only argument `value_wo` is available to use in place of `value`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments).
