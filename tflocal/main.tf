module "tflocal" {
  source                       = "app.terraform.io/hashicorp-v2/tflocal-cloud/aws"
  version                      = "0.8.0"
  tflocal_cloud                = "true"
  git_branch                   = var.git_branch
  tfe_ref                      = var.tfe_ref
  ngrok_domain                 = var.ngrok_domain
  ngrok_tunnel_token           = var.ngrok_tunnel_token
  private_github_token         = var.private_github_token
  private_github_user          = var.private_github_user
  gem_contribsys_key           = var.gem_contribsys_key
  ejson_file_name              = var.ejson_file_name
  ejson_file_content           = var.ejson_file_content
  artifactory_username         = var.artifactory_username
  artifactory_token            = var.artifactory_token
  quay_username                = var.quay_username
  quay_password                = var.quay_password
  artifactory_token            = var.artifactory_token
  artifactory_username         = var.artifactory_username
  datadog_api_key              = var.datadog_api_key
  run_cleanup_script           = var.run_cleanup_script
  env                          = var.env

  tags = {
    Codebase  = "hashicorp/terraform-provider-tfe"
    Purpose   = "terraform-provider-tfe integration tests"
    Workspace = terraform.workspace
  }
}
