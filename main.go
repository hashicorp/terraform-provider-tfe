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

	// Strip the leading log package prefix so hclog
	// can set the appropriate log level
	logFlags := log.Flags()
	logFlags &^= (log.Ldate | log.Ltime)
	log.SetFlags(logFlags)

	// terraform-plugin-mux here is used to combine multiple Terraform providers
	// built using different SDK and frameworks in order to combine them into a
	// single logical provider for Terraform to work with.
	// Here, we use one provider (tfe.Provider) that relies on the standard
	// terraform-plugin-sdk, and this is the main framework for used in this
	// provider. The second provider (tfe.PluginProviderServer) relies on the
	// lower level terraform-plugin-go to handle far more complex behavior, and
	// only should be used for functionality that is not present in the
	// common terraform-plugin- sdk framework.
	mux, err := tfmux.NewSchemaServerFactory(
		ctx, tfe.Provider().GRPCProvider, tfe.PluginProviderServer,
	)
	if err != nil {
		log.Printf("[ERROR] Could not setup a NewSchemaServerFactory using the providers: %v", err)
		os.Exit(1)
	}

	err = tf5server.Serve(tfeProviderName, func() tfprotov5.ProviderServer {
		return mux.Server()
	})
	if err != nil {
		log.Printf("[ERROR] Could not start serving the ProviderServer: %v", err)
		os.Exit(1)
	}
}
