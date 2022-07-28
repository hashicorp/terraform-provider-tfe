resource "tfe_terraform_version" "test" {
  version = "1.1.2-custom"
  url = "https://tfe-host.com/path/to/terraform.zip"
  sha = "e75ac73deb69a6b3aa667cb0b8b731aee79e2904"
}