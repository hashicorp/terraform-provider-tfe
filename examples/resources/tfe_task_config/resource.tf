resource "tfe_organization_run_task" "example" {
  organization = "my-org-name"
  url          = "https://external.service.com"
  name         = "example-task"
  enabled      = true
  description  = "An example run task"
}

resource "tfe_task_config" "example" {
  owner             = "my-org-name"
  task_id           = tfe_organization_run_task.example.id
  enabled           = true
  enforcement_level = "advisory"
  stages            = ["pre_plan", "post_plan"]
}
