// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"net/http"
	"testing"
)

func TestLoggingNewLoggingTransport_IsRoundTripper(t *testing.T) {
	transport := NewLoggingTransport("example", &http.Transport{})
	var _ http.RoundTripper = transport
}
