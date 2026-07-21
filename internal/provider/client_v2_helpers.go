// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/go-tfe/v2/api/models"
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
