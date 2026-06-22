terraform {
  cloud {
    organization = "foo"
    workspaces {
      name = "self-managed"
    }
  }
}

data "tfe_workspace" "self" {
  name         = "self-managed"
  organization = "foo"
}

resource "tfe_workspace_settings" "self" {
  workspace_id        = data.tfe_workspace.self.id
  assessments_enabled = true
  tags = {
    prod = "true"
  }
}
