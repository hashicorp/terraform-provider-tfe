// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider"
)

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

	var serveOpts []tf6server.ServeOpt

	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}
	// terraform-plugin-mux here is used to combine multiple Terraform providers
	// built using different SDK and frameworks in order to combine them into a
	// single logical provider for Terraform to work with.
	// - The classic provider relies on terraform-plugin-sdk, and has the bulk
	//   of the resources and data sources.
	// - The "next" provider relies on the newer terraform-plugin-framework, and
	//   we expect to migrate resources and data sources to it over time.
	// - The low-level provider relies on terraform-plugin-go to handle more
	//   complex behavior, and should only be used for functionality that is not
	//   available otherwise. We suspect the framework can supplant it, but have
	//   not proven that out yet.

	classicProvider := provider.Provider().GRPCProvider
	upgradedClassicProvider, err := tf5to6server.UpgradeServer(
		ctx,
		classicProvider,
	)
	nextProvider := providerserver.NewProtocol6(provider.NewFrameworkProvider())

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return upgradedClassicProvider
		},
		nextProvider,
	}

	mux, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Printf("[ERROR] Could not setup a mux server using the internal providers: %v", err)
		os.Exit(1)
	}

	err = tf6server.Serve(
		tfeProviderName,
		mux.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Printf("[ERROR] Could not start serving the ProviderServer: %v", err)
		os.Exit(1)
	}
}
