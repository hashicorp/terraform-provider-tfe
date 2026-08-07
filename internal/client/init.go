// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"sync"

	tfe "github.com/hashicorp/go-tfe"
	tfev2 "github.com/hashicorp/go-tfe/v2"
)

var clientCache *ClientConfigMap

func init() {
	clientCache = &ClientConfigMap{
		valuesV1: make(map[string]*tfe.Client),
		values:   make(map[string]*tfev2.Client),
		mu:       sync.Mutex{},
	}
}
