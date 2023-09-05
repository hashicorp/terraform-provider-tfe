// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"sync"

	tfe "github.com/hashicorp/go-tfe"
)

var clientCache *ClientConfigMap

func init() {
	clientCache = &ClientConfigMap{
		values: make(map[string]*tfe.Client),
		mu:     sync.Mutex{},
	}
}
