// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"io"

	tfe "github.com/hashicorp/go-tfe"
)

type workspaceNamesKey struct {
	organization, workspace string
}

type mockWorkspaces struct {
	options        testClientOptions
	workspaceNames map[workspaceNamesKey]*tfe.Workspace
}

// newMockWorkspaces creates a mock workspaces implementation. Any created
// workspaces will have the id given in defaultWorkspaceID.
func newMockWorkspaces(options testClientOptions) *mockWorkspaces {
	return &mockWorkspaces{
		options:        options,
		workspaceNames: make(map[workspaceNamesKey]*tfe.Workspace),
	}
}

var _ tfe.Workspaces = (*mockWorkspaces)(nil)

func (m *mockWorkspaces) List(ctx context.Context, organization string, options *tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Create(ctx context.Context, organization string, options tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
	ws := &tfe.Workspace{
		ID:   m.options.defaultWorkspaceID,
		Name: *options.Name,
		Organization: &tfe.Organization{
			Name: organization,
		},
		Permissions: &tfe.WorkspacePermissions{},
	}

	m.workspaceNames[workspaceNamesKey{organization, *options.Name}] = ws

	return ws, nil
}

func (m *mockWorkspaces) Read(ctx context.Context, organization, workspace string) (*tfe.Workspace, error) {
	w := m.workspaceNames[workspaceNamesKey{organization, workspace}]
	if w == nil {
		return nil, tfe.ErrResourceNotFound
	}

	return w, nil
}

func (m *mockWorkspaces) ReadWithOptions(ctx context.Context, organization, workspace string, options *tfe.WorkspaceReadOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) ReadByIDWithOptions(ctx context.Context, workspaceID string, options *tfe.WorkspaceReadOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Readme(ctx context.Context, workspaceID string) (io.Reader, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) ReadByID(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	for _, workspace := range m.workspaceNames {
		if workspace.ID == workspaceID {
			return workspace, nil
		}
	}
	return nil, tfe.ErrResourceNotFound
}

func (m *mockWorkspaces) Update(ctx context.Context, organization, workspace string, options tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) UpdateByID(ctx context.Context, workspaceID string, options tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Delete(ctx context.Context, organization, workspace string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) DeleteByID(ctx context.Context, workspaceID string) error {
	for key, workspace := range m.workspaceNames {
		if workspace.ID == workspaceID {
			delete(m.workspaceNames, key)
			return nil
		}
	}
	return fmt.Errorf("no workspace found with id %s", workspaceID)
}

func (m *mockWorkspaces) RemoveVCSConnection(ctx context.Context, organization, workspace string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveVCSConnectionByID(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Lock(ctx context.Context, workspaceID string, options tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Unlock(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) ForceUnlock(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) AssignSSHKey(ctx context.Context, workspaceID string, options tfe.WorkspaceAssignSSHKeyOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) UnassignSSHKey(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) ListRemoteStateConsumers(ctx context.Context, workspaceID string, options *tfe.RemoteStateConsumersListOptions) (*tfe.WorkspaceList, error) {
	if m.options.remoteStateConsumersResponse == "404" {
		return nil, tfe.ErrResourceNotFound
	} else if m.options.remoteStateConsumersResponse == "500" {
		return nil, errors.New("something is broken")
	}

	return &tfe.WorkspaceList{Items: []*tfe.Workspace{{ID: "ws-456"}}, Pagination: &tfe.Pagination{CurrentPage: 1, TotalPages: 1}}, nil
}

func (m *mockWorkspaces) AddRemoteStateConsumers(ctx context.Context, workspaceID string, options tfe.WorkspaceAddRemoteStateConsumersOptions) error {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveRemoteStateConsumers(ctx context.Context, workspaceID string, options tfe.WorkspaceRemoveRemoteStateConsumersOptions) error {
	panic("not implemented")
}

func (m *mockWorkspaces) UpdateRemoteStateConsumers(ctx context.Context, workspaceID string, options tfe.WorkspaceUpdateRemoteStateConsumersOptions) error {
	panic("not implemented")
}

func (m *mockWorkspaces) ListTags(ctx context.Context, workspaceID string, options *tfe.WorkspaceTagListOptions) (*tfe.TagList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) AddTags(ctx context.Context, workspaceID string, options tfe.WorkspaceAddTagsOptions) error {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveTags(ctx context.Context, workspaceID string, options tfe.WorkspaceRemoveTagsOptions) error {
	panic("not implemented")
}

func (m *mockWorkspaces) SafeDelete(ctx context.Context, organization string, workspace string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) SafeDeleteByID(ctx context.Context, workspaceID string) error {
	panic("not implemented")
}
