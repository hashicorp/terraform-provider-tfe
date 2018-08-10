package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-tfe/tfe"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tfe.Provider})
}
