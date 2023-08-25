// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/golang/mock/gomock"
	tfe "github.com/hashicorp/go-tfe"
	tfemocks "github.com/hashicorp/go-tfe/mocks"
)

func MockOrganizationMemberships(t *testing.T, client *tfe.Client, orgName string, organizationMemberships []*tfe.OrganizationMembership) {
	ctrl := gomock.NewController(t)
	mockOrganizationMembershipAPI := tfemocks.NewMockOrganizationMemberships(ctrl)
	organizationMembershipsList := tfe.OrganizationMembershipList{
		Items: organizationMemberships,
		Pagination: &tfe.Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			TotalCount:  len(organizationMemberships),
		},
	}

	mockOrganizationMembershipAPI.
		EXPECT().
		List(gomock.Any(), orgName, gomock.Any()).
		Return(&organizationMembershipsList, nil).
		AnyTimes()

	mockOrganizationMembershipAPI.
		EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, tfe.ErrInvalidOrg).
		AnyTimes()

	client.OrganizationMemberships = mockOrganizationMembershipAPI
}

func TestFetchOrganizationMembers(t *testing.T) {
	orgName := "hashicorp"

	tests := map[string]struct {
		members                []*tfe.OrganizationMembership
		org                    string
		err                    bool
		expectedMembers        []map[string]string
		expectedMembersWaiting []map[string]string
	}{
		"with non exisiting organization": {
			[]*tfe.OrganizationMembership{},
			"not-an-org",
			true,
			nil,
			nil,
		},
		"with no members": {
			[]*tfe.OrganizationMembership{},
			orgName,
			false,
			nil,
			nil,
		},
		"with both active and invited members": {
			activeAndInvitedOrganizationMemberships(orgName),
			orgName,
			false,
			[]map[string]string{{"user_id": "user-orgmember-1", "organization_membership_id": "ou-orgmember-1"}},
			[]map[string]string{{"user_id": "user-orgmember-2", "organization_membership_id": "ou-orgmember-2"}},
		},
	}

	client := testTfeClient(t, testClientOptions{defaultOrganization: orgName})

	for name, test := range tests {
		// Mock the Organization Membership
		MockOrganizationMemberships(t, client, orgName, test.members)
		t.Run(name, func(t *testing.T) {
			receivedMembers, receivedMembersWaiting, err := fetchOrganizationMembers(client, test.org)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			checkIsEqualMembers(t, receivedMembers, test.expectedMembers)
			checkIsEqualMembers(t, receivedMembersWaiting, test.expectedMembersWaiting)
		})
	}
}

func checkIsEqualMembers(t *testing.T, receivedMembers []map[string]string, expectedMembers []map[string]string) {
	if expectedMembers != nil && receivedMembers != nil {
		if len(expectedMembers) != len(receivedMembers) {
			t.Fatalf("wrong result\ngot: %#v\nwant: %#v", receivedMembers, expectedMembers)
		}

		// the test case only have 1 active and invited member
		if receivedMembers[0]["user_id"] != expectedMembers[0]["user_id"] || receivedMembers[0]["organization_membership_id"] != expectedMembers[0]["organization_membership_id"] {
			t.Fatalf("wrong result\ngot: %#v\nwant: %#v", receivedMembers[0], expectedMembers[0])
		}
	} else if (expectedMembers == nil && receivedMembers != nil) || (expectedMembers != nil && receivedMembers == nil) {
		t.Fatalf("wrong result\ngot: %#v\nwant: %#v", receivedMembers, expectedMembers)
	}
}

func activeAndInvitedOrganizationMemberships(orgName string) []*tfe.OrganizationMembership {
	return []*tfe.OrganizationMembership{
		{
			ID:           "ou-orgmember-1",
			Status:       tfe.OrganizationMembershipActive,
			Email:        "org_member_1@hashicorp.com",
			Organization: &tfe.Organization{Name: orgName},
			User: &tfe.User{
				ID: "user-orgmember-1",
			},
		},
		{
			ID:           "ou-orgmember-2",
			Status:       tfe.OrganizationMembershipInvited,
			Email:        "org_member_2@hashicorp.com",
			Organization: &tfe.Organization{Name: orgName},
			User: &tfe.User{
				ID: "user-orgmember-2",
			},
		},
	}
}
