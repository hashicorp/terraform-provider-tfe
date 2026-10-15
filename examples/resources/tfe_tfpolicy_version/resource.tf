# Requires a Terraform Enterprise administrator token.
# Replace the URLs and SHA-256 checksums with those of your binary builds.
resource "tfe_tfpolicy_version" "example" {
  version = "0.1.0-custom"
  enabled = true

  archs = [
    {
      os   = "linux"
      arch = "amd64"
      url  = "https://example.com/tfpolicy-plugin/0.1.0-custom/tfpolicy-plugin_0.1.0-custom_linux_amd64.zip"
      sha  = "5d34931df2bcc9fcf19f0023b4f97fad4679ffb87bee72d888c5cf9542ef3ec0"
    },
    {
      os   = "linux"
      arch = "arm64"
      url  = "https://example.com/tfpolicy-plugin/0.1.0-custom/tfpolicy-plugin_0.1.0-custom_linux_arm64.zip"
      sha  = "d5b1dafbe700d4c6f5620324d826ebac67c829f8da46a9ae868974153cbe13c1"
    },
  ]
}
