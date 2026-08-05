// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"testing"

	tfe "github.com/hashicorp/go-tfe/v2"
	"github.com/stretchr/testify/assert"
)

func TestFormatV2Error(t *testing.T) {
	t.Run("plain error", func(t *testing.T) {
		assert.Equal(t, "plain error", formatV2Error(errors.New("plain error")))
	})

	t.Run("API details", func(t *testing.T) {
		err := &tfe.APIError{StatusCode: 422, Message: "Unprocessable Entity", Details: []string{"first", "second"}}
		assert.Equal(t, "422 Unprocessable Entity: first; second", formatV2Error(err))
	})
}
