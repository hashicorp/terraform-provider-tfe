// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// computeWOHash returns a hex-encoded SHA-256 hash of the given value.
func computeWOHash(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
}

// privateStateSetter is satisfied by the Private field on Create/Update responses.
type privateStateSetter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// storeWOHash JSON-encodes the SHA-256 hash of woValue and stores it in private state under hashKey.
// Does nothing if woValue is null.
func storeWOHash(ctx context.Context, private privateStateSetter, hashKey string, woValue types.String, diags *diag.Diagnostics) {
	if woValue.IsNull() {
		// Clear any stale hash so that re-adding the same value later is treated as a new value.
		diags.Append(private.SetKey(ctx, hashKey, nil)...)
		return
	}
	hashJSON, err := json.Marshal(computeWOHash(woValue.ValueString()))
	if err != nil {
		diags.AddError("Failed to encode "+hashKey, err.Error())
		return
	}
	diags.Append(private.SetKey(ctx, hashKey, hashJSON)...)
}

// storeWOHashIfAutoManaged stores the hash of woValue only when the version attribute
// was not explicitly set in config (auto-managed mode).
func storeWOHashIfAutoManaged(ctx context.Context, private privateStateSetter, hashKey string, woValue types.String, configVersion types.Int64, diags *diag.Diagnostics) {
	if !configVersion.IsNull() {
		return
	}
	storeWOHash(ctx, private, hashKey, woValue, diags)
}

// blockWOToPlaintextTransition errors if the state has an active write-only version (woVersionAttr
// is non-null) while the plan sets the corresponding plaintext attribute (plaintextAttr is non-null).
func blockWOToPlaintextTransition(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, woVersionAttr, plaintextAttr string) {
	var stateVersion types.Int64
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(woVersionAttr), &stateVersion)...)
	if resp.Diagnostics.HasError() || stateVersion.IsNull() {
		return
	}

	var planPlaintext types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(plaintextAttr), &planPlaintext)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planPlaintext.IsNull() {
		resp.Diagnostics.AddError(
			"Cannot switch from write-only to plaintext",
			fmt.Sprintf("The %q attribute is currently managed as write-only. Setting %q would store the value in state, potentially exposing a previously secret value. Continue using the write-only attribute instead.", woVersionAttr, plaintextAttr),
		)
	}
}

// modifyPlanWOVersion manages the auto-detection version for a write-only attribute.
// If the version attribute is explicitly set in config (manual mode), no auto-detection is performed.
func modifyPlanWOVersion(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	woAttr, versionAttr, hashKey string,
) {
	// If version is explicitly set in config, use manual mode — skip auto-detection
	var configVersion types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(versionAttr), &configVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !configVersion.IsNull() {
		return
	}

	// Get write-only value from config
	var woValue types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(woAttr), &woValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if woValue.IsNull() {
		// Write-only value not set — clear the version
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Null())...)
		return
	}

	// On create (no prior state), set initial version to 1
	if req.State.Raw.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Value(1))...)
		return
	}

	var stateVersion types.Int64
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(versionAttr), &stateVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	currentVersion := int64(0)
	if !stateVersion.IsNull() && !stateVersion.IsUnknown() {
		currentVersion = stateVersion.ValueInt64()
	}

	if woValue.IsUnknown() {
		// Can't hash-compare an unknown value — assume it changed
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Value(currentVersion+1))...)
		return
	}

	newHash := computeWOHash(woValue.ValueString())

	// Compare new hash against stored hash in private state
	storedHashBytes, diags := req.Private.GetKey(ctx, hashKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storedHash string
	if storedHashBytes != nil {
		if err := json.Unmarshal(storedHashBytes, &storedHash); err != nil {
			resp.Diagnostics.AddError("Failed to decode "+woAttr+" hash", err.Error())
			return
		}
	}

	if !bytes.Equal([]byte(newHash), []byte(storedHash)) {
		// Hash changed — increment version
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(versionAttr), types.Int64Value(currentVersion+1))...)
	}
}
