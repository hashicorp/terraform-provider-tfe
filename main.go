package main

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	tf5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"
	tfmux "github.com/hashicorp/terraform-plugin-mux"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-provider-tfe/tfe"
)

func main() {
	ctx := context.Background()
	mainProvider := tfe.Provider().GRPCProvider
	altProvider := tfe.AltServer

	muxed, err := tfmux.NewSchemaServerFactory(ctx, mainProvider, altProvider)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	err = tf5server.Serve("registry.terraform.io/hashicorp/tfe", func() tfprotov5.ProviderServer {
		return muxed.Server()
	})
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
