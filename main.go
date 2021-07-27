package main

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	tf5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"
	tfmux "github.com/hashicorp/terraform-plugin-mux"
	"github.com/hashicorp/terraform-provider-tfe/tfe"
)

const (
	tfeProviderName = "registry.terraform.io/hashicorp/tfe"
)

func main() {
	ctx := context.Background()
	mux, err := tfmux.NewSchemaServerFactory(
		ctx, tfe.Provider().GRPCProvider, tfe.PluginProviderServer,
	)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	err = tf5server.Serve(tfeProviderName, func() tfprotov5.ProviderServer {
		return mux.Server()
	})
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
