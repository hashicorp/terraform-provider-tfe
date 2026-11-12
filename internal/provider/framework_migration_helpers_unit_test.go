// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"reflect"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestDiffStateResources(t *testing.T) {
	v0, v1 := uint64(0), uint64(1)

	base := func() *tfjson.StateResource {
		return &tfjson.StateResource{
			SchemaVersion:         1,
			AttributeValues:       map[string]any{"id": "x-1", "name": "a", "description": "", "tags": []any{}},
			IdentitySchemaVersion: &v0,
			IdentityValues:        map[string]any{"id": "x-1"},
		}
	}

	cases := map[string]struct {
		mutate func(*tfjson.StateResource)
		ignore []string
		want   []string
	}{
		"identical": {},
		"empty string becomes null": {
			mutate: func(r *tfjson.StateResource) { r.AttributeValues["description"] = nil },
			want:   []string{`attribute "description": "" -> <nil>`},
		},
		"empty list becomes null": {
			mutate: func(r *tfjson.StateResource) { r.AttributeValues["tags"] = nil },
			want:   []string{`attribute "tags": []interface {}{} -> <nil>`},
		},
		"attribute added and removed": {
			mutate: func(r *tfjson.StateResource) {
				delete(r.AttributeValues, "name")
				r.AttributeValues["extra"] = true
			},
			want: []string{
				`attribute "extra": absent -> true`,
				`attribute "name": "a" -> absent`,
			},
		},
		"ignored attribute": {
			mutate: func(r *tfjson.StateResource) { r.AttributeValues["name"] = "b" },
			ignore: []string{"name"},
		},
		"schema version": {
			mutate: func(r *tfjson.StateResource) { r.SchemaVersion = 2 },
			want:   []string{"schema_version: 1 -> 2"},
		},
		"identity": {
			mutate: func(r *tfjson.StateResource) {
				r.IdentitySchemaVersion = &v1
				r.IdentityValues["hostname"] = "app.terraform.io"
			},
			want: []string{
				"identity_schema_version: 0 -> 1",
				`identity "hostname": absent -> "app.terraform.io"`,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			after := base()
			if tc.mutate != nil {
				tc.mutate(after)
			}
			got := diffStateResources(base(), after, tc.ignore)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}
