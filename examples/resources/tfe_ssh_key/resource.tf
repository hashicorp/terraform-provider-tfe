resource "tfe_ssh_key" "test" {
  name         = "my-ssh-key-name"
  organization = "my-org-name"
  key          = "private-ssh-key"
}