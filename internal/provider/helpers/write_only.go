// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWriteOnlyValueStore(private PrivateState, attributeName string) *WriteOnlyValueStore {
	return &WriteOnlyValueStore{
		private:       private,
		attributeName: attributeName,
	}
}

type WriteOnlyValueStore struct {
	private       PrivateState
	attributeName string
}

type PrivateState interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

// MatchesPriorValue determines if the given string value matches the prior
// value in state by comparing the hased values of each.
func (w *WriteOnlyValueStore) MatchesPriorValue(ctx context.Context, configValue types.String) (bool, diag.Diagnostics) {
	serializedPriorValue, diags := w.private.GetKey(ctx, w.attributeName)
	var hashedPriorValue string
	err := json.Unmarshal(serializedPriorValue, &hashedPriorValue)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to unmarshal prior value for `%s`", w.attributeName), err.Error())
	}

	hashedValue := generateSHA256Hash(configValue.ValueString())
	return hashedPriorValue == hashedValue, diags
}

// PriorValueExists determines if a hashed prior value exists in state.
func (w *WriteOnlyValueStore) PriorValueExists(ctx context.Context) (bool, diag.Diagnostics) {
	serializedPriorValue, diags := w.private.GetKey(ctx, w.attributeName)
	return len(serializedPriorValue) != 0, diags
}

// SetPriorValue stores the hashed value of the given string value in state.
func (w *WriteOnlyValueStore) SetPriorValue(ctx context.Context, configValue types.String) diag.Diagnostics {
	// If not write-only, then remove the hashed value from private state.
	// Setting a key with an empty byte slice is interpreted by the framework as a request to remove the key from the ProviderData map.
	if configValue.IsNull() {
		return w.private.SetKey(ctx, w.attributeName, []byte(""))
	}

	// Store the hashed value of the string in private state.
	hashedValue := generateSHA256Hash(configValue.ValueString())
	return w.private.SetKey(ctx, w.attributeName, fmt.Appendf(nil, `"%s"`, hashedValue))
}

func generateSHA256Hash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
