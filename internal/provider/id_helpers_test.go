// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsResourceIDFormat(t *testing.T) {
	assert.Truef(t, isResourceIDFormat("tst", "tst-qwertyuiopasdfgh"), "correct ID format: all letters")
	assert.Truef(t, isResourceIDFormat("tst", "tst-1234567891234567"), "correct ID format: all numbers")
	assert.Truef(t, isResourceIDFormat("tst", "tst-QWERTYUPASDFGHJK"), "correct ID format: all uppercase")
	assert.Truef(t, isResourceIDFormat("tst", "tst-1q2W3e4R5t6Y7u89"), "correct ID format: alphanumeric")

	assert.Falsef(t, isResourceIDFormat("tst", "tst-only15charsXXXX"), "incorrect ID format: too short")
	assert.Falsef(t, isResourceIDFormat("tst", "tst-17charsLongXXXXXX"), "incorrect ID format: too long")
	assert.Falsef(t, isResourceIDFormat("tst", "1234567891234567"), "incorrect ID format: prefix missing")
	assert.Falsef(t, isResourceIDFormat("tst", "foo-1234567891234567"), "incorrect ID format: prefix incorrect")
	assert.Falsef(t, isResourceIDFormat("tst", "tst-1234567890123456"), "incorrect ID format: contains a 0")
	assert.Falsef(t, isResourceIDFormat("tst", "tst-QWERTYUOPASDFGHJ"), "incorrect ID format: contains a O")
	assert.Falsef(t, isResourceIDFormat("tst", "tst-QWERTYUIPASDFGHJ"), "incorrect ID format: contains a I")
	assert.Falsef(t, isResourceIDFormat("tst", "tst-asdfghjklzxcvbnm"), "incorrect ID format: contains a l")
	assert.Falsef(t, isResourceIDFormat("tst", "^[[-qwertyuiopasdfgh"), "incorrect ID format: prefix has regex cahrs")
}
