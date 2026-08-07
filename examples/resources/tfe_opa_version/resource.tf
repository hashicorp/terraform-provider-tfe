# Basic usage

resource "tfe_opa_version" "test" {
  version = "0.58.0-custom"
  url     = "https://tfe-host.com/path/to/opa"
  sha     = "e75ac73deb69a6b3aa667cb0b8b731aee79e2904"
}
