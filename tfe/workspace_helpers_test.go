package tfe

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

type mockWorkspaceReader struct{}

func (*mockWorkspaceReader) Read(ctx context.Context, organization string, workspace string) (*tfe.Workspace, error) {
	if organization == "hashicorp" && workspace == "a-workspace" {
		return &tfe.Workspace{
			ID: "ws-123",
		}, nil
	}

	return nil, tfe.ErrResourceNotFound
}

func TestFetchWorkspaceExternalID(t *testing.T) {
	tests := map[string]struct {
		def  string
		want string
		err  bool
	}{
		"non exisiting organization": {
			"not-an-org/workspace",
			"",
			true,
		},
		"non exisiting workspace": {
			"hashicorp/not-a-workspace",
			"",
			true,
		},
		"found workspace": {
			"hashicorp/a-workspace",
			"ws-123",
			false,
		},
	}

	reader := &mockWorkspaceReader{}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchWorkspaceExternalID(test.def, reader)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if got != test.want {
				t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, test.want)
			}
		})
	}
}

type mockWorkspaceIDReader struct{}

func (*mockWorkspaceIDReader) ReadByID(ctx context.Context, id string) (*tfe.Workspace, error) {
	if id == "ws-123" {
		return &tfe.Workspace{
			Name: "a-workspace",
			Organization: &tfe.Organization{
				Name: "hashicorp",
			},
		}, nil
	}

	return nil, tfe.ErrResourceNotFound
}

func TestFetchWorkspaceHumanID(t *testing.T) {
	tests := map[string]struct {
		def  string
		want string
		err  bool
	}{
		"non exisiting workspace": {
			"ws-notathing",
			"",
			true,
		},
		"found workspace": {
			"ws-123",
			"hashicorp/a-workspace",
			false,
		},
	}

	reader := &mockWorkspaceIDReader{}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchWorkspaceHumanID(test.def, reader)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if got != test.want {
				t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, test.want)
			}
		})
	}
}

func TestPackWorkspaceID(t *testing.T) {
	cases := []struct {
		w   *tfe.Workspace
		id  string
		err bool
	}{
		{
			w: &tfe.Workspace{
				Name: "my-workspace-name",
				Organization: &tfe.Organization{
					Name: "my-org-name",
				},
			},
			id:  "my-org-name/my-workspace-name",
			err: false,
		},
		{
			w: &tfe.Workspace{
				Name: "my-workspace-name",
			},
			id:  "",
			err: true,
		},
	}

	for _, tc := range cases {
		id, err := packWorkspaceID(tc.w)
		if (err != nil) != tc.err {
			t.Fatalf("expected error is %t, got %v", tc.err, err)
		}

		if tc.id != id {
			t.Fatalf("expected ID %q, got %q", tc.id, id)
		}
	}
}

func TestUnpackWorkspaceID(t *testing.T) {
	cases := []struct {
		id   string
		org  string
		name string
		err  bool
	}{
		{
			id:   "my-org-name/my-workspace-name",
			org:  "my-org-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "my-workspace-name|my-org-name",
			org:  "my-org-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "some-invalid-id",
			org:  "",
			name: "",
			err:  true,
		},
	}

	for _, tc := range cases {
		org, name, err := unpackWorkspaceID(tc.id)
		if (err != nil) != tc.err {
			t.Fatalf("expected error is %t, got %v", tc.err, err)
		}

		if tc.org != org {
			t.Fatalf("expected organization %q, got %q", tc.org, org)
		}

		if tc.name != name {
			t.Fatalf("expected name %q, got %q", tc.name, name)
		}
	}
}
