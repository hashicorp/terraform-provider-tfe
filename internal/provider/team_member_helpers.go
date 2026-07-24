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

// teamMembersListUsersV2 returns the users of a team, mirroring go-tfe v1's
// TeamMembers.List/ListUsers (GET /teams/:id?include=users).
func teamMembersListUsersV2(ctx context.Context, api *v2api.ApiClient, teamID string) ([]models.Usersable, error) {
	include := teamitem.USERS_GETINCLUDEQUERYPARAMETERTYPE
	result, err := api.Teams().ById(teamID).Get(ctx, withQueryParams(&teams.ItemRequestBuilderGetQueryParameters{
		Include: &include,
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
	var users []models.Usersable
	for _, ref := range relationships.GetUsers().GetData() {
		if user := findIncludedUser(included, valueOrZero(ref.GetId())); user != nil {
			users = append(users, user)
		}
	}
	return users, nil
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
