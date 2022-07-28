# To import a variable that's part of a workspace, use <ORGANIZATION NAME>/<WORKSPACE NAME>/<VARIABLE ID> as the import ID. For example:

terraform import tfe_variable.test my-org-name/varset-47qC3LmA47piVan7/var-5rTwnSaRPogw6apb

# To import a variable that's part of a variable set, use <ORGANIZATION NAME>/<VARIABLE SET ID>/<VARIABLE ID> as the import ID. For example:

terraform import tfe_variable.test my-org-name/my-workspace-name/var-5rTwnSaRPogw6apb
