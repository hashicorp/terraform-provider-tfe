resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@example.com"
}

resource "tfe_organization_run_task" "example" {
  organization = tfe_organization.test.name
  url          = "https://external.service.com"
  name         = "test-task"
}

# Apply the task to every workspace in the organization
resource "tfe_task_config" "global" {
  task_id           = tfe_organization_run_task.example.id
  global            = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
  organization      = tfe_organization.test.name
  owner             = tfe_organization.test.name
}
