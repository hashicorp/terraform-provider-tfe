resource "tfe_organization_run_task" "example" {
  organization = "my-org-name"
  url          = "https://external.service.com"
  name         = "task-name"
}

# Apply the task to every workspace in the organization
resource "tfe_task_config" "global" {
  task_id           = tfe_organization_run_task.example.id
  global            = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
}

# Apply the task to the workspaces of specific projects
resource "tfe_project" "testproject" {
  organization = "my-org-name"
  name         = "test-project-name"
}

resource "tfe_task_config" "projects" {
  task_id           = tfe_organization_run_task.example.id
  enforcement_level = "advisory"
  stages            = ["pre_plan", "post_plan"]
  project_ids       = [tfe_project.testproject.id]
}
