// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/jsonapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var workspaceIDRegexp = regexp.MustCompile("^ws-[a-zA-Z0-9]{16}$")

func validTagName(tag string) bool {
	tagPattern := regexp.MustCompile(`\A[a-z0-9](?:[a-z0-9_:-]*[a-z0-9])?\z`)
	return tagPattern.MatchString(tag)
}

func hasConfiguredTriggerConflict(prefixes types.List, patterns types.List) bool {
	return !prefixes.IsNull() && !prefixes.IsUnknown() && !patterns.IsNull() && !patterns.IsUnknown()
}

func flattenAutoDestroyAt(a jsonapi.NullableAttr[time.Time]) (*string, error) {
	if !a.IsSpecified() {
		return nil, nil
	}

	autoDestroyTime, err := a.Get()
	if err != nil {
		return nil, err
	}

	autoDestroyAt := autoDestroyTime.Format(time.RFC3339)
	return &autoDestroyAt, nil
}

func safeWorkspaceDelete(ctx context.Context, config ConfiguredClient, id string) error {
	return retry.RetryContext(ctx, time.Duration(5)*time.Minute, func() *retry.RetryError {
		err := config.Client.Workspaces.SafeDeleteByID(ctx, id)
		if errors.Is(err, tfe.ErrWorkspaceStillProcessing) {
			return retry.RetryableError(err)
		} else if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

func errWorkspaceSafeDeleteWithPermission(workspaceID string, err error) error {
	if err != nil {
		if strings.HasPrefix(err.Error(), "conflict") {
			return fmt.Errorf("error deleting workspace %s: %w\nThis workspace may either have managed resources in state or has a latest state that's still being processed. Add force_delete = true to the resource config to delete this workspace", workspaceID, err)
		}
		return err
	}
	return nil
}

func errWorkspaceResourceCountCheck(workspaceID string, resourceCount int) error {
	if resourceCount > 0 {
		return fmt.Errorf(
			"error deleting workspace %s: This workspace has %v resources under management and must be force deleted by setting force_delete = true", workspaceID, resourceCount)
	}
	return nil
}

func expandWorkspaceTagBindings(tags types.Map) []*tfe.TagBinding {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}

	out := make([]*tfe.TagBinding, 0, len(tags.Elements()))
	for key, v := range tags.Elements() {
		if strVal, ok := v.(types.String); ok {
			out = append(out, &tfe.TagBinding{Key: key, Value: strVal.ValueString()})
		}
	}
	return out
}

func expandWorkspaceStringList(ctx context.Context, v types.List, diags *diag.Diagnostics) ([]string, bool) {
	if v.IsNull() || v.IsUnknown() {
		return nil, false
	}

	items := []string{}
	diags.Append(v.ElementsAs(ctx, &items, false)...)
	return items, true
}

func expandWorkspaceTagNames(ctx context.Context, v types.Set, diags *diag.Diagnostics) ([]*tfe.Tag, bool) {
	if v.IsNull() || v.IsUnknown() {
		return nil, false
	}

	names := []string{}
	diags.Append(v.ElementsAs(ctx, &names, false)...)
	tags := make([]*tfe.Tag, 0, len(names))
	for _, name := range names {
		tags = append(tags, &tfe.Tag{Name: name})
	}
	return tags, true
}

func expandWorkspaceVCSRepoOptions(ctx context.Context, v types.List, diags *diag.Diagnostics, includeOptionalEmpty bool) *tfe.VCSRepoOptions {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	var vcs []modelWorkspaceVCSRepo
	diags.Append(v.ElementsAs(ctx, &vcs, false)...)
	if len(vcs) == 0 {
		return nil
	}

	repo := vcs[0]
	out := &tfe.VCSRepoOptions{
		Identifier:        tfe.String(repo.Identifier.ValueString()),
		IngressSubmodules: tfe.Bool(repo.IngressSubmodules.ValueBool()),
		TagsRegex:         tfe.String(repo.TagsRegex.ValueString()),
	}

	if includeOptionalEmpty {
		out.Branch = tfe.String(repo.Branch.ValueString())
		out.OAuthTokenID = tfe.String(repo.OAuthTokenID.ValueString())
		out.GHAInstallationID = tfe.String(repo.GithubAppInstallationID.ValueString())
		return out
	}

	if !repo.Branch.IsNull() && repo.Branch.ValueString() != "" {
		out.Branch = tfe.String(repo.Branch.ValueString())
	}
	if !repo.OAuthTokenID.IsNull() && repo.OAuthTokenID.ValueString() != "" {
		out.OAuthTokenID = tfe.String(repo.OAuthTokenID.ValueString())
	}
	if !repo.GithubAppInstallationID.IsNull() && repo.GithubAppInstallationID.ValueString() != "" {
		out.GHAInstallationID = tfe.String(repo.GithubAppInstallationID.ValueString())
	}

	return out
}

func stringToFramework(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func stringSliceToList(v []string) types.List {
	vals := make([]attr.Value, 0, len(v))
	for _, s := range v {
		vals = append(vals, types.StringValue(s))
	}
	return types.ListValueMust(types.StringType, vals)
}

func stringSliceToSet(v []string) types.Set {
	vals := make([]attr.Value, 0, len(v))
	for _, s := range v {
		vals = append(vals, types.StringValue(s))
	}
	return types.SetValueMust(types.StringType, vals)
}

func mapTypeFromStringMap(m map[string]interface{}) types.Map {
	vals := make(map[string]attr.Value, len(m))
	for k, v := range m {
		vals[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.MapValueMust(types.StringType, vals)
}

func mapFromStringMapType(m types.Map) map[string]interface{} {
	result := map[string]interface{}{}
	if m.IsNull() || m.IsUnknown() {
		return result
	}
	for k, v := range m.Elements() {
		if sv, ok := v.(types.String); ok {
			result[k] = sv.ValueString()
		}
	}
	return result
}

func boolValueOrDefault(v types.Bool, def bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueBool()
}

func stringValueOrDefault(v types.String, def string) string {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueString()
}
