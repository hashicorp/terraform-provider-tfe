variable "ssh_key" {
  type      = string
  ephemeral = true
}

resource "tfe_ssh_key" "test" {
  name           = "my-ssh-key-name"
  organization   = "my-org-name"
  key_wo         = var.ssh_key
  key_wo_version = 1
}
