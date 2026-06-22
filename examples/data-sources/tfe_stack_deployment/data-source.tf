data "tfe_stack_deployment" "staging" {
  organization = "my-example-org"
  name         = "staging"
  stack        = "example-stack"
}
