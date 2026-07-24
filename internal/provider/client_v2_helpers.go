// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/go-tfe/v2/api/models"
	abstractions "github.com/microsoft/kiota-abstractions-go"
)

// valueOrZero dereferences an optional pointer returned by a go-tfe v2
// generated getter, returning the type's zero value when the pointer is nil.
// This preserves go-tfe v1 behavior, where absent JSON:API attributes were
// left as zero values instead of nil pointers.
func valueOrZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// nextPageNumber returns the next page number from a go-tfe v2 list response
// pagination object, or nil when there are no more pages. Callers should
// follow next-page rather than relying on total-pages/total-count, which some
// endpoints omit.
func nextPageNumber(p models.Paginationable) *int32 {
	if p == nil {
		return nil
	}
	return p.GetNextPage()
}

// paginationMetaGetter is satisfied by every go-tfe v2 generated list
// response's `meta` type (e.g. ItemSshKeysGetResponse_meta,
// ItemTeamsGetResponse_meta): each is a distinct generated type, but all of
// them implement GetPagination() identically.
type paginationMetaGetter interface {
	GetPagination() models.Paginationable
}

// nextPageFromMeta returns the next page number from a go-tfe v2 list
// response's `meta` field, or nil when meta or pagination is absent or there
// are no more pages.
func nextPageFromMeta(meta paginationMetaGetter) *int32 {
	if meta == nil {
		return nil
	}
	return nextPageNumber(meta.GetPagination())
}

// withQueryParams wraps a go-tfe v2 generated query-parameters struct in the
// kiota-abstractions RequestConfiguration envelope that every generated
// request builder's Get/List method expects.
func withQueryParams[T any](queryParams *T) *abstractions.RequestConfiguration[T] {
	return &abstractions.RequestConfiguration[T]{QueryParameters: queryParams}
}

// includedUserGetter is satisfied by every go-tfe v2 generated `included`
// array element type (e.g. the composed-type wrappers for both the
// organization-memberships list and single-item responses): each is a
// distinct generated type, but all of them implement GetUsers() identically.
type includedUserGetter interface {
	GetUsers() models.Usersable
}

// findIncludedUser returns the full user record with the given ID from a
// JSON:API `included` array, or nil when it is not present. Requires the
// go-tfe v2 client to correctly discriminate composed `included` array
// elements by their JSON:API `type`.
func findIncludedUser[T includedUserGetter](included []T, userID string) models.Usersable {
	if userID == "" {
		return nil
	}
	for _, record := range included {
		if user := record.GetUsers(); user != nil && valueOrZero(user.GetId()) == userID {
			return user
		}
	}
	return nil
}
