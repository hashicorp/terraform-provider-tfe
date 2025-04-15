// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	tfe "github.com/hashicorp/go-tfe"
)

type capabilitiesResolver interface {
	IsCloud() bool
	RemoteTFEVersion() string
}

func newDefaultCapabilityResolver(client *tfe.Client) capabilitiesResolver {
	return &defaultCapabilityResolver{
		client: client,
	}
}

type defaultCapabilityResolver struct {
	client *tfe.Client
}

func (r *defaultCapabilityResolver) IsCloud() bool {
	return r.client.IsCloud()
}

func (r *defaultCapabilityResolver) RemoteTFEVersion() string {
	return r.client.RemoteTFEVersion()
}
