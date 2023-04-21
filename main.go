// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-provider-tfe/tfe"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

const (
	tfeProviderName = "registry.terraform.io/hashicorp/tfe"
)

func main() {
	ctx := context.Background()

	// Remove any date and time prefix in log package function output to
	// prevent duplicate timestamp and incorrect log level setting
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()

	var serveOpts []tf5server.ServeOpt

	if *debugFlag {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
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
	mux, err := tf5muxserver.NewMuxServer(
		ctx, tfe.Provider().GRPCProvider, tfe.PluginProviderServer,
	)
	if err != nil {
		log.Printf("[ERROR] Could not setup a NewSchemaServerFactory using the providers: %v", err)
		os.Exit(1)
	}

	err = tf5server.Serve(tfeProviderName, mux.ProviderServer, serveOpts...)
	if err != nil {
		log.Printf("[ERROR] Could not start serving the ProviderServer: %v", err)
		os.Exit(1)
	}
}
