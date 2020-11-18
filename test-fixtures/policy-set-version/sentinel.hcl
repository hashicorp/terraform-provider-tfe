policy "a-policy" {
  source = "./a-policy.sentinel"
  enforcement_level = "soft-mandatory"
}

module "a-module" {
  source = "./modules/a-module.sentinel"
}
