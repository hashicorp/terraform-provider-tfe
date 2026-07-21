// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// membershipsHandler serves GET
// /api/v2/organizations/{orgName}/organization-memberships with the given
// JSON:API pages (served in order by the page[number] query parameter), and
// responds 404 for any other organization.
func membershipsHandler(orgName string, pages map[string]string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/"+orgName+"/organization-memberships", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page[number]")
		if page == "" {
			page = "1"
		}
		body, ok := pages[page]
		if !ok {
			http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, body)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
	})
	return mux
}

func membershipResource(membershipID, userID, email, status string) string {
	return fmt.Sprintf(`{
		"id": %q,
		"type": "organization-memberships",
		"attributes": {"email": %q, "status": %q},
		"relationships": {
			"user": {"data": {"id": %q, "type": "users"}},
			"organization": {"data": {"id": "hashicorp", "type": "organizations"}}
		}
	}`, membershipID, email, status, userID)
}

func paginationMeta(currentPage int, nextPage, totalPages string) string {
	return fmt.Sprintf(`{"pagination": {"current-page": %d, "next-page": %s, "total-pages": %s}}`, currentPage, nextPage, totalPages)
}

func TestFetchOrganizationMembers(t *testing.T) {
	orgName := "hashicorp"

	singlePage := map[string]string{
		"1": fmt.Sprintf(`{"data": [%s, %s], "meta": %s}`,
			membershipResource("ou-orgmember-1", "user-orgmember-1", "org_member_1@hashicorp.com", "active"),
			membershipResource("ou-orgmember-2", "user-orgmember-2", "org_member_2@hashicorp.com", "invited"),
			paginationMeta(1, "null", "1"),
		),
	}

	multiPage := map[string]string{
		"1": fmt.Sprintf(`{"data": [%s], "meta": %s}`,
			membershipResource("ou-orgmember-1", "user-orgmember-1", "org_member_1@hashicorp.com", "active"),
			paginationMeta(1, "2", "2"),
		),
		"2": fmt.Sprintf(`{"data": [%s], "meta": %s}`,
			membershipResource("ou-orgmember-2", "user-orgmember-2", "org_member_2@hashicorp.com", "invited"),
			paginationMeta(2, "null", "2"),
		),
	}

	emptyPage := map[string]string{
		"1": fmt.Sprintf(`{"data": [], "meta": %s}`, paginationMeta(1, "null", "1")),
	}

	tests := map[string]struct {
		pages                  map[string]string
		org                    string
		err                    bool
		expectedMembers        []map[string]string
		expectedMembersWaiting []map[string]string
	}{
		"with non existing organization": {
			emptyPage,
			"not-an-org",
			true,
			nil,
			nil,
		},
		"with no members": {
			emptyPage,
			orgName,
			false,
			nil,
			nil,
		},
		"with both active and invited members": {
			singlePage,
			orgName,
			false,
			[]map[string]string{{"user_id": "user-orgmember-1", "organization_membership_id": "ou-orgmember-1"}},
			[]map[string]string{{"user_id": "user-orgmember-2", "organization_membership_id": "ou-orgmember-2"}},
		},
		"with members across multiple pages": {
			multiPage,
			orgName,
			false,
			[]map[string]string{{"user_id": "user-orgmember-1", "organization_membership_id": "ou-orgmember-1"}},
			[]map[string]string{{"user_id": "user-orgmember-2", "organization_membership_id": "ou-orgmember-2"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := testTfeClientV2(t, membershipsHandler(orgName, test.pages))

			receivedMembers, receivedMembersWaiting, err := fetchOrganizationMembers(client, test.org)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			checkIsEqualMembers(t, receivedMembers, test.expectedMembers)
			checkIsEqualMembers(t, receivedMembersWaiting, test.expectedMembersWaiting)
		})
	}
}

func userResource(userID, email, username string) string {
	return fmt.Sprintf(`{
		"id": %q,
		"type": "users",
		"attributes": {"email": %q, "username": %q}
	}`, userID, email, username)
}

// TestFetchOrganizationMemberByNameOrEmailV2 is a regression test for a
// go-tfe v2 generated-client bug: the composed-type factories for JSON:API
// `included` array elements did not discriminate on the `type` field, so
// GetUsers() on an included record was always nil regardless of what was
// actually included. That bug is why tfe_organization_membership stayed on
// go-tfe v1 for a time. It's fixed upstream (a discriminator was added to the
// OpenAPI schema for the `included` arrays on the organization-memberships
// endpoints), and this test exercises real JSON:API deserialization end to
// end (not just in-memory structs) so a regression would be caught here
// without requiring a live TFE org or CI-seeded users.
func TestFetchOrganizationMemberByNameOrEmailV2(t *testing.T) {
	orgName := "hashicorp"

	pageFor := func(memberships, included string) map[string]string {
		return map[string]string{
			"1": fmt.Sprintf(`{"data": [%s], "included": [%s], "meta": %s}`, memberships, included, paginationMeta(1, "null", "1")),
		}
	}

	matching := pageFor(
		membershipResource("ou-orgmember-1", "user-orgmember-1", "org_member_1@hashicorp.com", "active"),
		userResource("user-orgmember-1", "org_member_1@hashicorp.com", "orgmember1"),
	)

	noResults := map[string]string{
		"1": fmt.Sprintf(`{"data": [], "meta": %s}`, paginationMeta(1, "null", "1")),
	}

	tests := map[string]struct {
		pages        map[string]string
		username     string
		email        string
		expectExists bool
		expectedID   string
		err          bool
	}{
		"neither username nor email": {
			matching,
			"",
			"",
			false,
			"",
			true,
		},
		"matching email": {
			matching,
			"",
			"org_member_1@hashicorp.com",
			true,
			"ou-orgmember-1",
			false,
		},
		"matching username": {
			matching,
			"orgmember1",
			"",
			true,
			"ou-orgmember-1",
			false,
		},
		"single result with mismatched identity": {
			matching,
			"someone-else",
			"",
			false,
			"",
			true,
		},
		"no results": {
			noResults,
			"orgmember1",
			"",
			false,
			"",
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := testTfeClientV2(t, membershipsHandler(orgName, test.pages))

			got, err := fetchOrganizationMemberByNameOrEmailV2(context.Background(), client, orgName, test.username, test.email)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if test.expectExists {
				if got == nil || valueOrZero(got.GetId()) != test.expectedID {
					t.Fatalf("wrong result\ngot: %#v\nwant membership with ID %q", got, test.expectedID)
				}
			} else if got != nil {
				t.Fatalf("wrong result\ngot: %#v\nwant: nil", got)
			}
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
