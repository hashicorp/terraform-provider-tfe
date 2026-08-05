// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/stretchr/testify/assert"
)

func TestFormatV2Error(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Empty(t, formatV2Error(nil))
	})

	t.Run("plain error", func(t *testing.T) {
		assert.Equal(t, "plain error", formatV2Error(errors.New("plain error")))
	})

	t.Run("API error without details", func(t *testing.T) {
		err := &tfe.APIError{StatusCode: 404, Message: "Not Found"}
		assert.Equal(t, "404 Not Found", formatV2Error(err))
	})

	t.Run("API details", func(t *testing.T) {
		err := &tfe.APIError{StatusCode: 422, Message: "Unprocessable Entity", Details: []string{"first", "second"}}
		assert.Equal(t, "422 Unprocessable Entity: first; second", formatV2Error(err))
	})

	t.Run("wrapped API details", func(t *testing.T) {
		err := fmt.Errorf("create failed: %w", &tfe.APIError{StatusCode: 422, Message: "Unprocessable Entity", Details: []string{"invalid URL"}})
		assert.Equal(t, "create failed: 422 Unprocessable Entity: invalid URL", formatV2Error(err))
	})
}
