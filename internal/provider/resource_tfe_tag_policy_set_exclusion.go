// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &resourceTFETagPolicySetExclusion{}
	_ resource.ResourceWithConfigure   = &resourceTFETagPolicySetExclusion{}
	_ resource.ResourceWithImportState = &resourceTFETagPolicySetExclusion{}
)

type resourceTFETagPolicySetExclusion struct {
	config ConfiguredClient
}

func NewTagPolicySetExclusionResource() resource.Resource {
	return &resourceTFETagPolicySetExclusion{}
}

type modelTagPolicySetExclusion struct {
	ID          types.String `tfsdk:"id"`
	PolicySetID types.String `tfsdk:"policy_set_id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
}

func (r *resourceTFETagPolicySetExclusion) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(ConfiguredClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected tfe.ConfiguredClient, got %T. This is a bug in the tfe provider, so please report it on GitHub.", req.ProviderData),
		)
		return
	}

	r.config = client
}

// Metadata implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_policy_set_exclusion"
}

// Schema implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a resource which manages tag exclusions on a policy set.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The composite ID of the tag exclusion, in the format <POLICY_SET_ID>/<TAG_KEY> or <POLICY_SET_ID>/<TAG_KEY>/<TAG_VALUE>.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "The ID of the policy set to add the tag exclusion to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^polset-[a-zA-Z0-9]{16}$`),
						"must be a valid policy set ID (e.g. polset-<RANDOM_STRING>)",
					),
				},
			},
			"key": schema.StringAttribute{
				Description: "The tag key for the exclusion.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "The tag value for the tag exclusion. If not set, this becomes a key-only tag and only matches workspaces that also have a key-only tag with the given key.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

// Create implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan modelTagPolicySetExclusion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetID := plan.PolicySetID.ValueString()
	key := plan.Key.ValueString()
	valuePtr := plan.Value.ValueStringPointer()

	tflog.Debug(ctx, fmt.Sprintf("Adding tag exclusion (key=%s, value=%s) to policy set %s", key, ptrValueOrNil(valuePtr), policySetID))
	err := r.config.Client.PolicySets.AddTagSelectors(ctx, policySetID, tfe.PolicySetAddTagSelectorsOptions{
		TagSelectors: []*tfe.PolicySetTagSelector{
			{Key: key, Value: valuePtr, IsExclude: true},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Tag Exclusion to Policy Set",
			fmt.Sprintf("An error was encountered when adding tag exclusion (key=%q, value=%q) to policy set %q: %s", key, ptrValueOrNil(valuePtr), policySetID, err),
		)
		return
	}

	if valuePtr != nil {
		plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", policySetID, key, *valuePtr))
	} else {
		plan.ID = types.StringValue(fmt.Sprintf("%s/%s", policySetID, key))
	}

	tflog.Debug(ctx, fmt.Sprintf("Creation of tag exclusion (key=%s, value=%s) for policy set %s is complete", key, ptrValueOrNil(valuePtr), policySetID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state modelTagPolicySetExclusion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetID := state.PolicySetID.ValueString()
	key := state.Key.ValueString()
	valuePtr := state.Value.ValueStringPointer()

	tflog.Debug(ctx, fmt.Sprintf("Reading tag exclusion (key=%s, value=%s) from policy set %s", key, ptrValueOrNil(valuePtr), policySetID))
	policySet, err := r.config.Client.PolicySets.Read(ctx, policySetID)
	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", policySetID))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Policy Set",
			fmt.Sprintf("An error was encountered when reading policy set %q: %s", policySetID, err),
		)
		return
	}

	for _, ts := range policySet.TagSelectors {
		if ts.Key == key && ts.IsExclude && r.tagValueMatches(ts.Value, state.Value) {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Tag exclusion (key=%s, value=%s) not found in policy set %s. Removing from state.", key, ptrValueOrNil(valuePtr), policySetID))
	resp.State.RemoveResource(ctx)
}

// Update implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This method is a no-op but required by the framework
	var plan modelTagPolicySetExclusion
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements [resource.Resource].
func (r *resourceTFETagPolicySetExclusion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelTagPolicySetExclusion
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetID := state.PolicySetID.ValueString()
	key := state.Key.ValueString()
	valuePtr := state.Value.ValueStringPointer()

	tflog.Debug(ctx, fmt.Sprintf("Removing tag exclusion (key=%s, value=%s) from policy set (%s)", key, ptrValueOrNil(valuePtr), policySetID))
	err := r.config.Client.PolicySets.RemoveTagSelectors(ctx, policySetID, tfe.PolicySetRemoveTagSelectorsOptions{
		TagSelectors: []*tfe.PolicySetTagSelector{
			{Key: key, Value: valuePtr, IsExclude: true},
		},
	})

	if err != nil && errors.Is(err, tfe.ErrResourceNotFound) {
		tflog.Debug(ctx, fmt.Sprintf("Policy set %s no longer exists.", policySetID))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Tag Exclusion from Policy Set",
			fmt.Sprintf("An error was encountered when removing tag exclusion (key=%q, value=%q) from policy set %q: %s", key, ptrValueOrNil(valuePtr), policySetID, err),
		)
		return
	}
}

// ImportState implements [resource.ResourceWithImportState].
func (r *resourceTFETagPolicySetExclusion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	splitID := strings.SplitN(req.ID, "/", 3)
	if len(splitID) < 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID Format",
			fmt.Sprintf("The import ID must be in the format <POLICY_SET_ID>/<TAG_KEY> or <POLICY_SET_ID>/<TAG_KEY>/<TAG_VALUE>. Got: %q", req.ID),
		)
		return
	}

	policySetID := splitID[0]
	tagKey := splitID[1]

	matched, _ := regexp.MatchString(`^polset-[a-zA-Z0-9]{16}$`, policySetID)
	if !matched {
		resp.Diagnostics.AddError(
			"Invalid Policy Set ID",
			fmt.Sprintf("The policy set ID %q is not valid. Expected format: polset-<16 alphanumeric chars>.", policySetID),
		)
		return
	}

	var tagValue *string
	if len(splitID) == 3 {
		v := splitID[2]
		tagValue = &v
	}

	tflog.Debug(ctx, fmt.Sprintf("Importing tag exclusion (key=%s, value=%s) for policy set %s", tagKey, ptrValueOrNil(tagValue), policySetID))

	policySet, err := r.config.Client.PolicySets.Read(ctx, policySetID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Policy Set",
			fmt.Sprintf("An error was encountered when reading policy set %q: %s", policySetID, err),
		)
		return
	}

	for _, ts := range policySet.TagSelectors {
		if ts.Key != tagKey || !ts.IsExclude {
			continue
		}
		if !r.tagValueMatches(ts.Value, types.StringPointerValue(tagValue)) {
			continue
		}

		var id string
		if ts.Value != nil {
			id = fmt.Sprintf("%s/%s/%s", policySetID, tagKey, *ts.Value)
		} else {
			id = fmt.Sprintf("%s/%s", policySetID, tagKey)
		}

		state := modelTagPolicySetExclusion{
			ID:          types.StringValue(id),
			PolicySetID: types.StringValue(policySetID),
			Key:         types.StringValue(tagKey),
			Value:       types.StringPointerValue(ts.Value),
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.Diagnostics.AddError(
		"Tag Exclusion Not Found",
		fmt.Sprintf("Tag exclusion (key=%q, value=%q) not found in policy set %q.", tagKey, ptrValueOrNil(tagValue), policySetID),
	)
}

// tagValueMatches returns true when the tag value from the API
// matches the value stored in state. A null stateValue means the tag has no
// value (key-only), which corresponds to a nil API value.
func (r *resourceTFETagPolicySetExclusion) tagValueMatches(tsValue *string, stateValue types.String) bool {
	if stateValue.IsNull() {
		return tsValue == nil
	}
	return tsValue != nil && *tsValue == stateValue.ValueString()
}
