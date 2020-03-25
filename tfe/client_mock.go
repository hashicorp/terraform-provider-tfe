package tfe

import (
	"context"

	tfe "github.com/hashicorp/go-tfe"
)

type workspaceNamesKey struct {
	organization, workspace string
}

type mockWorkspaces struct {
	workspaceNames map[workspaceNamesKey]*tfe.Workspace
}

func newMockWorkspaces() *mockWorkspaces {
	return &mockWorkspaces{
		workspaceNames: make(map[workspaceNamesKey]*tfe.Workspace),
	}
}

func (m *mockWorkspaces) List(ctx context.Context, organization string, options tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Create(ctx context.Context, organization string, options tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
	ws := &tfe.Workspace{
		ID:   options.ID,
		Name: *options.Name,
		Organization: &tfe.Organization{
			Name: organization,
		},
	}

	m.workspaceNames[workspaceNamesKey{organization, *options.Name}] = ws

	return ws, nil
}

func (m *mockWorkspaces) Read(ctx context.Context, organization string, workspace string) (*tfe.Workspace, error) {
	w := m.workspaceNames[workspaceNamesKey{organization, workspace}]
	if w == nil {
		return nil, tfe.ErrResourceNotFound
	}

	return w, nil
}

func (m *mockWorkspaces) ReadByID(ctx context.Context, workspaceID string) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Update(ctx context.Context, organization string, workspace string, options tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) UpdateByID(ctx context.Context, workspaceID string, options tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Delete(ctx context.Context, organization string, workspace string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) DeleteByID(ctx context.Context, workspaceID string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveVCSConnection(ctx context.Context, organization string, workspace string) (*tfe.Workspace, error) {
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
