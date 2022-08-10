data "tfe_outputs" "foo" {
  organization = "my-org"
  workspace = "my-workspace"
}

resource "random_id" "vpc_id" {
  keepers = {
    # Generate a new ID any time the value of 'bar' in workspace 'my-org/my-workspace' changes.
    bar = data.tfe_outputs.foo.values.bar
  }

  byte_length = 8
}