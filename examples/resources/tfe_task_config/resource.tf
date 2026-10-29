resource "tfe_organization_run_task" "example" {
  organization = "org-name"
  url          = "https://external.service.com"
  name         = "test-task"
}

# Apply the task to every workspace in the organization
resource "tfe_task_config" "global" {
  task_id           = tfe_organization_run_task.example.id
  global            = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
}
