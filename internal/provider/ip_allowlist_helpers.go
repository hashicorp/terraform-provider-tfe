// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	tfev2 "github.com/hashicorp/go-tfe/v2"
	cidrrangelists "github.com/hashicorp/go-tfe/v2/api/cidrrangelists"
	tfev2models "github.com/hashicorp/go-tfe/v2/api/models"
	organizations "github.com/hashicorp/go-tfe/v2/api/organizations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

// requestConfigurationCidrRanges is a convenience alias for the request
// configuration accepted by the list-CIDR-ranges endpoint.
type requestConfigurationCidrRanges = abstractions.RequestConfiguration[cidrrangelists.ItemRelationshipsCidrRangesRequestBuilderGetQueryParameters]

// requestConfigurationOrgCidrRangeLists is a convenience alias for the request
// configuration accepted by the organization CIDR-range-lists endpoint.
type requestConfigurationOrgCidrRangeLists = abstractions.RequestConfiguration[organizations.ItemCidrRangeListsRequestBuilderGetQueryParameters]

// IP allowlists are referred to as "CIDR range lists" in the HCP Terraform API
// and go-tfe v2 client. The enforcement_scope values below map to the API enum.
const (
	ipAllowlistScopeOrganization       = "organization"
	ipAllowlistScopeAllAgentPools      = "all_agent_pools"
	ipAllowlistScopeSelectedAgentPools = "selected_agent_pools"
)

// modelTFECIDRRange is the model for a single CIDR range nested within an IP
// allowlist. The range identifier assigned by the API is intentionally not
// exposed; ranges are reconciled by their CIDR value.
type modelTFECIDRRange struct {
	Range       types.String `tfsdk:"range"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// cidrRangeAttrTypes describes the attribute types of a single cidr_range
// element. It is used when converting API results back into a types.Set.
var cidrRangeAttrTypes = map[string]attr.Type{
	"range":       types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
}

func cidrRangeObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: cidrRangeAttrTypes}
}

// isV2ResourceNotFound reports whether the given go-tfe v2 client error
// represents an HTTP 404 Not Found response.
func isV2ResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *tfev2models.Errors
	if errors.As(err, &apiErr) {
		return apiErr.ResponseStatusCode == http.StatusNotFound
	}
	return false
}

// enforcementScopeToV2 converts the Terraform enforcement_scope string into the
// go-tfe v2 enum value.
func enforcementScopeToV2(scope string) (tfev2models.CidrRangeLists_attributes_enforcementScope, error) {
	switch scope {
	case ipAllowlistScopeOrganization:
		return tfev2models.ORGANIZATION_CIDRRANGELISTS_ATTRIBUTES_ENFORCEMENTSCOPE, nil
	case ipAllowlistScopeAllAgentPools:
		return tfev2models.ALL_AGENT_POOLS_CIDRRANGELISTS_ATTRIBUTES_ENFORCEMENTSCOPE, nil
	case ipAllowlistScopeSelectedAgentPools:
		return tfev2models.SELECTED_AGENT_POOLS_CIDRRANGELISTS_ATTRIBUTES_ENFORCEMENTSCOPE, nil
	default:
		return 0, fmt.Errorf("invalid enforcement_scope %q", scope)
	}
}

// enforcementScopeFromV2 converts the go-tfe v2 enum value into the Terraform
// enforcement_scope string.
func enforcementScopeFromV2(scope *tfev2models.CidrRangeLists_attributes_enforcementScope) string {
	if scope == nil {
		return ""
	}
	return scope.String()
}

// agentPoolIDsBody builds the request body used to assign or unassign agent
// pools from a CIDR range list.
func agentPoolIDsBody(ids []string) tfev2models.AgentPoolIdsable {
	body := tfev2models.NewAgentPoolIds()
	poolType := tfev2models.AGENTPOOLS_AGENTPOOLIDS_DATA_TYPE
	data := make([]tfev2models.AgentPoolIds_dataable, 0, len(ids))
	for i := range ids {
		id := ids[i]
		d := tfev2models.NewAgentPoolIds_data()
		d.SetId(&id)
		d.SetTypeEscaped(&poolType)
		data = append(data, d)
	}
	body.SetData(data)
	return body
}

// nestedCidrRangeData builds a NestedCidrRange used when seeding CIDR ranges at
// creation time (embedded in the CIDR range list create request).
func nestedCidrRangeData(m modelTFECIDRRange) tfev2models.NestedCidrRangeable {
	attrs := tfev2models.NewNestedCidrRange_attributes()
	rng := m.Range.ValueString()
	enabled := m.Enabled.ValueBool()
	attrs.SetRangeEscaped(&rng)
	attrs.SetEnabled(&enabled)
	if !m.Description.IsNull() {
		desc := m.Description.ValueString()
		attrs.SetDescription(&desc)
	}

	rangeType := tfev2models.CIDRRANGES_NESTEDCIDRRANGE_TYPE
	data := tfev2models.NewNestedCidrRange()
	data.SetTypeEscaped(&rangeType)
	data.SetAttributes(attrs)
	return data
}

// cidrRangeEnvelope builds a CidrRangesEnvelope used to create or update an
// individual CIDR range.
func cidrRangeEnvelope(m modelTFECIDRRange) tfev2models.CidrRangesEnvelopeable {
	attrs := tfev2models.NewCidrRanges_attributes()
	rng := m.Range.ValueString()
	enabled := m.Enabled.ValueBool()
	attrs.SetRangeEscaped(&rng)
	attrs.SetEnabled(&enabled)
	if !m.Description.IsNull() {
		desc := m.Description.ValueString()
		attrs.SetDescription(&desc)
	}

	rangeType := tfev2models.CIDRRANGES_CIDRRANGES_TYPE
	data := tfev2models.NewCidrRanges()
	data.SetTypeEscaped(&rangeType)
	data.SetAttributes(attrs)

	envelope := tfev2models.NewCidrRangesEnvelope()
	envelope.SetData(data)
	return envelope
}

// currentAgentPoolIDs extracts the assigned agent pool IDs from a CIDR range
// list's relationships.
func currentAgentPoolIDs(list tfev2models.CidrRangeListsable) []string {
	if list == nil {
		return nil
	}
	rel := list.GetRelationships()
	if rel == nil {
		return nil
	}
	agentPools := rel.GetAgentPools()
	if agentPools == nil {
		return nil
	}
	ids := make([]string, 0, len(agentPools.GetData()))
	for _, d := range agentPools.GetData() {
		if d.GetId() != nil {
			ids = append(ids, *d.GetId())
		}
	}
	return ids
}

// readIPAllowlistRanges fetches every CIDR range that belongs to a CIDR range
// list, transparently following pagination.
func readIPAllowlistRanges(ctx context.Context, clientV2 *tfev2.Client, listID string) ([]tfev2models.CidrRangesable, error) {
	const pageSize int32 = 100

	var ranges []tfev2models.CidrRangesable
	pageNumber := int32(1)

	for {
		size := pageSize
		number := pageNumber
		cfg := &cidrrangelists.ItemRelationshipsCidrRangesRequestBuilderGetQueryParameters{
			Pagesize:   &size,
			Pagenumber: &number,
		}
		requestConfig := &requestConfigurationCidrRanges{QueryParameters: cfg}

		resp, err := clientV2.API.
			CidrRangeLists().
			ByCidr_range_list_id(listID).
			Relationships().
			CidrRanges().
			Get(ctx, requestConfig)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			break
		}

		data := resp.GetData()
		ranges = append(ranges, data...)

		if len(data) < int(pageSize) {
			break
		}
		pageNumber++
	}

	return ranges, nil
}

// cidrRangeSetFromAPI converts API CIDR ranges into a types.Set of cidr_range
// objects for storage in Terraform state.
func cidrRangeSetFromAPI(ctx context.Context, apiRanges []tfev2models.CidrRangesable) (types.Set, diag.Diagnostics) {
	models := make([]modelTFECIDRRange, 0, len(apiRanges))
	for _, r := range apiRanges {
		attrs := r.GetAttributes()
		if attrs == nil {
			continue
		}

		rng := ""
		if attrs.GetRangeEscaped() != nil {
			rng = *attrs.GetRangeEscaped()
		}
		description := types.StringNull()
		if attrs.GetDescription() != nil {
			description = types.StringValue(*attrs.GetDescription())
		}
		enabled := true
		if attrs.GetEnabled() != nil {
			enabled = *attrs.GetEnabled()
		}

		models = append(models, modelTFECIDRRange{
			Range:       types.StringValue(rng),
			Description: description,
			Enabled:     types.BoolValue(enabled),
		})
	}

	return types.SetValueFrom(ctx, cidrRangeObjectType(), models)
}

// stringSliceDifference returns the elements in a that are not present in b.
func stringSliceDifference(a, b []string) []string {
	inB := make(map[string]struct{}, len(b))
	for _, v := range b {
		inB[v] = struct{}{}
	}
	var diff []string
	for _, v := range a {
		if _, ok := inB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// setToStringSlice converts a types.Set of strings into a []string, appending
// any conversion errors to the supplied diagnostics.
func setToStringSlice(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(set.ElementsAs(ctx, &out, false)...)
	return out
}

// findIPAllowlistByName searches an organization's CIDR range lists for one with
// the given name, following pagination. It returns the matching list's ID and
// whether a match was found.
func findIPAllowlistByName(ctx context.Context, clientV2 *tfev2.Client, organization, name string) (string, bool, error) {
	const pageSize int32 = 100

	pageNumber := int32(1)
	for {
		size := pageSize
		number := pageNumber
		cfg := &organizations.ItemCidrRangeListsRequestBuilderGetQueryParameters{
			Pagesize:   &size,
			Pagenumber: &number,
		}
		requestConfig := &requestConfigurationOrgCidrRangeLists{QueryParameters: cfg}

		resp, err := clientV2.API.
			Organizations().
			ByOrganization_name(organization).
			CidrRangeLists().
			Get(ctx, requestConfig)
		if err != nil {
			return "", false, err
		}
		if resp == nil {
			return "", false, nil
		}

		data := resp.GetData()
		for _, list := range data {
			if list.GetAttributes() == nil || list.GetAttributes().GetName() == nil {
				continue
			}
			if *list.GetAttributes().GetName() == name && list.GetId() != nil {
				return *list.GetId(), true, nil
			}
		}

		if len(data) < int(pageSize) {
			break
		}
		pageNumber++
	}

	return "", false, nil
}
