// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	v2api "github.com/hashicorp/go-tfe/v2/api"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/go-tfe/v2/api/teams"
	teamitem "github.com/hashicorp/go-tfe/v2/api/teams/item"
)

// These helpers back go-tfe v1's TeamMembers.{Add,Remove,List,ListOrganizationMemberships}
// for the go-tfe v2 generated client. Team membership can be managed either
// by username (JSON:API type "users", with the username placed in the "id"
// field - a long-standing, intentional Atlas convention for this specific
// relationship endpoint) or by organization membership ID (JSON:API type
// "organization-memberships"). Both call graphs are used across the
// tfe_team_member, tfe_team_members, tfe_team_organization_member, and
// tfe_team_organization_members resources, so the add/remove/list operations
// are centralized here rather than duplicated per resource.

// buildUsersIdentifierDoc constructs a UsersIdentifierArrayDocument from
// a slice of usernames, following the Atlas convention of placing the username
// in the JSON:API "id" field for team-membership relationship endpoints.
func buildUsersIdentifierDoc(usernames []string) *models.UsersIdentifierArrayDocument {
	data := make([]models.UsersIdentifierArrayDocument_dataable, 0, len(usernames))
	for _, username := range usernames {
		item := models.NewUsersIdentifierArrayDocument_data()
		item.SetId(ptr(username))
		item.SetTypeEscaped(ptr(models.USERS_USERSIDENTIFIERARRAYDOCUMENT_DATA_TYPE))
		data = append(data, item)
	}
	doc := models.NewUsersIdentifierArrayDocument()
	doc.SetData(data)
	return doc
}

// teamMembersAddUsersV2 adds the given usernames to a team.
func teamMembersAddUsersV2(ctx context.Context, api *v2api.ApiClient, teamID string, usernames []string) error {
	if len(usernames) == 0 {
		return nil
	}
	return api.Teams().ById(teamID).Relationships().Users().Post(ctx, buildUsersIdentifierDoc(usernames), nil)
}

// teamMembersRemoveUsersV2 removes the given usernames from a team.
func teamMembersRemoveUsersV2(ctx context.Context, api *v2api.ApiClient, teamID string, usernames []string) error {
	if len(usernames) == 0 {
		return nil
	}
	return api.Teams().ById(teamID).Relationships().Users().Delete(ctx, buildUsersIdentifierDoc(usernames), nil)
}

// teamMembersListUsersV2 returns the usernames of a team's members,
// mirroring go-tfe v1's TeamMembers.List/ListUsers (GET
// /teams/:id?include=users).
//
// NOTE: The go-tfe v2 generated client mis-discriminates this endpoint's
// `included` array: its OpenAPI schema declares `anyOf` (rather than
// `oneOf`) for the `users`/`organization-memberships` composed type, and the
// generated discriminator function unconditionally tries to decode every
// included record as `organization-memberships` first, regardless of its
// actual JSON:API `type`. Because the two schemas share a JSON:API
// envelope, this "succeeds" for `users` records too: the `id` is decoded
// correctly, but `username` (and other user-only attributes) land in the
// decoded object's AdditionalData map instead of a typed field, and
// `included[].GetUsers()` is always nil. teamUsernameFromIncluded below
// works around this by falling back to that AdditionalData when the typed
// accessor comes up empty.
func teamMembersListUsersV2(ctx context.Context, api *v2api.ApiClient, teamID string) ([]string, error) {
	include := teamitem.USERS_GETINCLUDEQUERYPARAMETERTYPE
	result, err := api.Teams().ById(teamID).Get(ctx, withQueryParams(&teams.ItemRequestBuilderGetQueryParameters{
		Include: []teamitem.GetIncludeQueryParameterType{include},
	}))
	if err != nil {
		return nil, err
	}
	if result == nil || result.GetData() == nil {
		return nil, fmt.Errorf("no data returned reading team %s", teamID)
	}

	relationships := result.GetData().GetRelationships()
	if relationships == nil || relationships.GetUsers() == nil {
		return nil, nil
	}

	included := result.GetIncluded()
	var usernames []string
	for _, ref := range relationships.GetUsers().GetData() {
		userID := valueOrZero(ref.GetId())
		if userID == "" {
			continue
		}
		if username := teamUsernameFromIncluded(included, userID); username != "" {
			usernames = append(usernames, username)
		}
	}
	return usernames, nil
}

// teamUsernameFromIncluded finds the username of the user with the given ID
// in a team's `included` array. See the note on teamMembersListUsersV2 for
// why this can't simply rely on the composed type's typed GetUsers()
// accessor.
func teamUsernameFromIncluded(included []teams.ItemGetResponse_GetResponse_includedable, userID string) string {
	for _, record := range included {
		if user := record.GetUsers(); user != nil {
			if valueOrZero(user.GetId()) == userID {
				return valueOrZero(user.GetAttributes().GetUsername())
			}
			continue
		}

		// Work around the client's discriminator bug: a `users` record
		// that was mis-decoded as `organization-memberships` still has
		// the correct `id`, with `username` stashed in AdditionalData.
		om := record.GetOrganizationMemberships()
		if om == nil || valueOrZero(om.GetId()) != userID {
			continue
		}
		attrs := om.GetAttributes()
		if attrs == nil {
			continue
		}
		switch v := attrs.GetAdditionalData()["username"].(type) {
		case *string:
			if v != nil {
				return *v
			}
		case string:
			return v
		}
	}
	return ""
}

// buildOrgMembershipsIdentifierDoc constructs an
// OrganizationMembershipsIdentifierArrayDocument from a slice of membership IDs.
func buildOrgMembershipsIdentifierDoc(membershipIDs []string) *models.OrganizationMembershipsIdentifierArrayDocument {
	data := make([]models.OrganizationMembershipsIdentifierArrayDocument_dataable, 0, len(membershipIDs))
	for _, id := range membershipIDs {
		item := models.NewOrganizationMembershipsIdentifierArrayDocument_data()
		item.SetId(ptr(id))
		item.SetTypeEscaped(ptr(models.ORGANIZATIONMEMBERSHIPS_ORGANIZATIONMEMBERSHIPSIDENTIFIERARRAYDOCUMENT_DATA_TYPE))
		data = append(data, item)
	}
	doc := models.NewOrganizationMembershipsIdentifierArrayDocument()
	doc.SetData(data)
	return doc
}

// teamMembersAddOrgMembershipsV2 adds the given organization membership IDs
// to a team.
func teamMembersAddOrgMembershipsV2(ctx context.Context, api *v2api.ApiClient, teamID string, membershipIDs []string) error {
	if len(membershipIDs) == 0 {
		return nil
	}
	return api.Teams().ById(teamID).Relationships().OrganizationMemberships().Post(ctx, buildOrgMembershipsIdentifierDoc(membershipIDs), nil)
}

// teamMembersRemoveOrgMembershipsV2 removes the given organization
// membership IDs from a team.
func teamMembersRemoveOrgMembershipsV2(ctx context.Context, api *v2api.ApiClient, teamID string, membershipIDs []string) error {
	if len(membershipIDs) == 0 {
		return nil
	}
	return api.Teams().ById(teamID).Relationships().OrganizationMemberships().Delete(ctx, buildOrgMembershipsIdentifierDoc(membershipIDs), nil)
}

// teamMembersListOrgMembershipsV2 returns all organization memberships
// associated with a team, following pagination, mirroring go-tfe v1's
// TeamMembers.ListOrganizationMemberships.
func teamMembersListOrgMembershipsV2(ctx context.Context, api *v2api.ApiClient, teamID string) ([]models.OrganizationMembershipsable, error) {
	builder := api.Teams().ById(teamID).Relationships().OrganizationMemberships()

	pageSize := int32(100)
	queryParams := &teams.ItemRelationshipsOrganizationMembershipsRequestBuilderGetQueryParameters{
		Pagesize: &pageSize,
	}

	result, err := builder.Get(ctx, withQueryParams(queryParams))
	if err != nil {
		return nil, err
	}

	var memberships []models.OrganizationMembershipsable
	for {
		memberships = append(memberships, result.GetData()...)

		nextPage := nextPageFromMeta(result.GetMeta())
		if nextPage == nil {
			break
		}

		queryParams = &teams.ItemRelationshipsOrganizationMembershipsRequestBuilderGetQueryParameters{
			Pagesize:   &pageSize,
			Pagenumber: nextPage,
		}
		result, err = builder.Get(ctx, withQueryParams(queryParams))
		if err != nil {
			return nil, err
		}
	}

	return memberships, nil
}
