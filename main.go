package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-provider-tfe/tfe"
)

const (
	tfeProviderName = "registry.terraform.io/hashicorp/tfe"
	version         = "dev"
)

func main() {
	ctx := context.Background()

	// Remove any date and time prefix in log package function output to
	// prevent duplicate timestamp and incorrect log level setting
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()

	var serveOpts []tf6server.ServeOpt
	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	upgradedSDKServer, err := tfe.UpgradedProviderServer()
	if err != nil {
		log.Printf("[ERROR] Could not upgrade legacy server to protocol v6: %v", err)
		os.Exit(1)
	}

	upgradedPluginServer, err := tfe.UpgradedPluginProviderServer()
	if err != nil {
		log.Printf("[ERROR] Could not upgrade provider server to protocol v6: %v", err)
		os.Exit(1)
	}

	// terraform-plugin-mux here is used to combine multiple Terraform providers
	// built using different SDK and frameworks in order to combine them into a
	// single logical provider for Terraform to work with.
	// Here, we use one provider (tfe.Provider) that relies on the standard
	// terraform-plugin-sdk, and this is the main framework for used in this
	// provider. The second provider (tfe.PluginProviderServer) relies on the
	// lower level terraform-plugin-go to handle far more complex behavior, and
	// only should be used for functionality that is not present in the
	// common terraform-plugin- sdk framework.
	providers := []func() tfprotov6.ProviderServer{
		// The main, legacy sdk provider (terraform-plugin-sdk)
		func() tfprotov6.ProviderServer { return upgradedSDKServer },

		// Lower-level provider that defines the behavior for the tfe_outputs data source
		func() tfprotov6.ProviderServer { return upgradedPluginServer },

		// Newer, framework-based provider (terraform-framework-plugin)
		providerserver.NewProtocol6(tfe.NewFrameworkProvider(version)),
	}

	mux, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Printf("[ERROR] Could not setup a server using the providers: %v", err)
		os.Exit(1)
	}

	log.Printf("[DEBUG] provider name: %s", tfeProviderName)

	err = tf6server.Serve(tfeProviderName, mux.ProviderServer, serveOpts...)
	if err != nil {
		log.Printf("[ERROR] Could not start serving the ProviderServer: %v", err)
		os.Exit(1)
	}
}
