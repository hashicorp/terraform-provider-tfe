// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	tfemocks "github.com/hashicorp/go-tfe/mocks"
	"go.uber.org/mock/gomock"
)

func TestWaitForModuleVersion(t *testing.T) {
	moduleID := tfe.RegistryModuleID{Organization: "org", Namespace: "org", Name: "mod", Provider: "aws"}
	version := func(s tfe.RegistryModuleVersionStatus) *tfe.RegistryModuleVersion {
		return &tfe.RegistryModuleVersion{Version: "1.0.0", Status: s}
	}

	cases := map[string]struct {
		responses []*tfe.RegistryModuleVersion
		errs      []error
		wantErr   string
	}{
		"ok immediately": {
			responses: []*tfe.RegistryModuleVersion{version(tfe.RegistryModuleVersionStatusOk)},
			errs:      []error{nil},
		},
		"empty status treated as ready": {
			responses: []*tfe.RegistryModuleVersion{version("")},
			errs:      []error{nil},
		},
		"not found then ingesting then ok": {
			responses: []*tfe.RegistryModuleVersion{nil, version(tfe.RegistryModuleVersionStatusRegIngressing), version(tfe.RegistryModuleVersionStatusOk)},
			errs:      []error{tfe.ErrResourceNotFound, nil, nil},
		},
		"ingest failure stops immediately": {
			responses: []*tfe.RegistryModuleVersion{version(tfe.RegistryModuleVersionStatusRegIngressFailed)},
			errs:      []error{nil},
			wantErr:   `failed to ingest: status "reg_ingress_failed"`,
		},
		"clone failure stops immediately": {
			responses: []*tfe.RegistryModuleVersion{version(tfe.RegistryModuleVersionStatusCloneFailed)},
			errs:      []error{nil},
			wantErr:   `failed to ingest: status "clone_failed"`,
		},
		"other error stops immediately": {
			responses: []*tfe.RegistryModuleVersion{nil},
			errs:      []error{errors.New("boom")},
			wantErr:   "boom",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mock := tfemocks.NewMockRegistryModules(ctrl)
			var calls []any
			for i := range tc.responses {
				calls = append(calls, mock.EXPECT().
					ReadVersion(gomock.Any(), moduleID, "1.0.0").
					Return(tc.responses[i], tc.errs[i]))
			}
			gomock.InOrder(calls...)

			err := waitForModuleVersion(context.Background(), &tfe.Client{RegistryModules: mock}, moduleID, "1.0.0")
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}
