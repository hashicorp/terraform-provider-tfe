# With write-only HMAC key

variable "hmac_key" {
  description = "The HMAC key for use in the write only field"
}

resource "tfe_organization_run_task" "example" {
  organization        = "org-name"
  url                 = "https://external.service.com"
  name                = "task-name"
  enabled             = true
  description         = "An example task"
  hmac_key_wo         = var.hmac_key
  hmac_key_wo_version = 1
}
