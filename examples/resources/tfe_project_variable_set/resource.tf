resource "tfe_organization" "test" {
   name  = "my-org-name"
   email = "admin@company.com"
 }
 resource "tfe_project" "test" {
   name         = "my-project-name"
   organization = tfe_organization.test.name
 }
 resource "tfe_variable_set" "test" {
   name         = "Test Varset"
   description  = "Some description."
   organization = tfe_organization.test.name
 }
 resource "tfe_project_variable_set" "test" {
   variable_set_id = tfe_variable_set.test.id
   project_id      = tfe_project.test.id
 }