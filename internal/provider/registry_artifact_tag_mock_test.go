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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
	"go.uber.org/mock/gomock"
)

func TestRegistryArtifactTag_mockCreate(t *testing.T) {
	testCases := map[string]struct {
		artifactType string
		artifactID   string
		tags         map[string]string
		readErr      error
		updateErr    error
		wantSummary  string
	}{
		"module": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
			tags:         map[string]string{"env": "prod"},
		},
		"provider": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			tags:         map[string]string{"team": "platform"},
		},
		"component": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			tags:         map[string]string{"env": "prod", "team": "platform"},
		},
		"module not found": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-missing",
			tags:         map[string]string{"env": "prod"},
			readErr:      tfe.ErrResourceNotFound,
			wantSummary:  "Registry Artifact Not Found",
		},
		"module update error": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
			tags:         map[string]string{"env": "prod"},
			updateErr:    errors.New("update failed"),
			wantSummary:  "Error Adding Tags to Registry Artifact",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := &tfe.Client{}
			expectArtifactTagWrite(t, ctrl, client, artifactTagWrite{
				artifactType: testCase.artifactType,
				artifactID:   testCase.artifactID,
				tags:         testCase.tags,
				readErr:      testCase.readErr,
				updateErr:    testCase.updateErr,
			})

			r := &resourceTFERegistryArtifactTag{config: ConfiguredClient{Client: client}}
			sch := artifactTagResourceSchema(t, r)
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
			r.Create(ctx, resource.CreateRequest{
				Plan: mustArtifactTagPlan(t, sch, artifactTagModel(testCase.artifactType, testCase.artifactID, "", testCase.tags)),
			}, resp)

			if testCase.wantSummary != "" {
				assertDiagnosticSummary(t, resp.Diagnostics, testCase.wantSummary)
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("create diagnostics: %v", resp.Diagnostics)
			}

			var got modelRegistryArtifactTag
			if diags := resp.State.Get(ctx, &got); diags.HasError() {
				t.Fatalf("state get: %v", diags)
			}
			wantID := testCase.artifactType + "/" + testCase.artifactID
			if got.ID.ValueString() != wantID {
				t.Fatalf("id = %q, want %q", got.ID.ValueString(), wantID)
			}
			if diff := diffRegistryArtifactTagMap(modelTagMap(got.Tags), testCase.tags); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestRegistryArtifactTag_mockUpdate(t *testing.T) {
	testCases := map[string]struct {
		artifactType string
		artifactID   string
		tags         map[string]string
		updateErr    error
		wantSummary  string
	}{
		"module": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
			tags:         map[string]string{"env": "staging"},
		},
		"provider": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			tags:         map[string]string{"team": "infra"},
		},
		"component": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			tags:         map[string]string{"env": "dev"},
		},
		"update error": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			tags:         map[string]string{"team": "infra"},
			updateErr:    errors.New("update failed"),
			wantSummary:  "Error Updating Tags on Registry Artifact",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := &tfe.Client{}
			expectArtifactTagWrite(t, ctrl, client, artifactTagWrite{
				artifactType: testCase.artifactType,
				artifactID:   testCase.artifactID,
				tags:         testCase.tags,
				updateErr:    testCase.updateErr,
			})

			r := &resourceTFERegistryArtifactTag{config: ConfiguredClient{Client: client}}
			sch := artifactTagResourceSchema(t, r)
			model := artifactTagModel(testCase.artifactType, testCase.artifactID, testCase.artifactType+"/"+testCase.artifactID, testCase.tags)
			resp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
			r.Update(ctx, resource.UpdateRequest{Plan: mustArtifactTagPlan(t, sch, model)}, resp)

			if testCase.wantSummary != "" {
				assertDiagnosticSummary(t, resp.Diagnostics, testCase.wantSummary)
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("update diagnostics: %v", resp.Diagnostics)
			}

			var got modelRegistryArtifactTag
			if diags := resp.State.Get(ctx, &got); diags.HasError() {
				t.Fatalf("state get: %v", diags)
			}
			if diff := diffRegistryArtifactTagMap(modelTagMap(got.Tags), testCase.tags); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestRegistryArtifactTag_mockDelete(t *testing.T) {
	testCases := map[string]struct {
		artifactType string
		artifactID   string
		readErr      error
		updateErr    error
		wantSummary  string
	}{
		"module clears tags": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
		},
		"provider clears tags": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
		},
		"component clears tags": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
		},
		"already gone": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-missing",
			readErr:      tfe.ErrResourceNotFound,
		},
		"update error": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			updateErr:    errors.New("delete failed"),
			wantSummary:  "Error Removing Tags from Registry Artifact",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := &tfe.Client{}
			expectArtifactTagWrite(t, ctrl, client, artifactTagWrite{
				artifactType: testCase.artifactType,
				artifactID:   testCase.artifactID,
				tags:         map[string]string{},
				readErr:      testCase.readErr,
				updateErr:    testCase.updateErr,
				emptySlice:   true,
			})

			r := &resourceTFERegistryArtifactTag{config: ConfiguredClient{Client: client}}
			sch := artifactTagResourceSchema(t, r)
			model := artifactTagModel(testCase.artifactType, testCase.artifactID, testCase.artifactType+"/"+testCase.artifactID, map[string]string{"env": "prod"})
			resp := &resource.DeleteResponse{}
			r.Delete(ctx, resource.DeleteRequest{State: mustArtifactTagState(t, sch, model)}, resp)

			if testCase.wantSummary != "" {
				assertDiagnosticSummary(t, resp.Diagnostics, testCase.wantSummary)
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("delete diagnostics: %v", resp.Diagnostics)
			}
		})
	}
}

func TestRegistryArtifactTag_mockRead(t *testing.T) {
	testCases := map[string]struct {
		artifactType string
		artifactID   string
		tags         map[string]string
		listErr      error
		wantSummary  string
		wantRemoved  bool
	}{
		"module": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
			tags:         map[string]string{"env": "prod"},
		},
		"provider": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			tags:         map[string]string{"team": "platform"},
		},
		"component": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			tags:         map[string]string{"env": "dev"},
		},
		"not found": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-missing",
			listErr:      tfe.ErrResourceNotFound,
			wantRemoved:  true,
		},
		"list error": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			listErr:      errors.New("list failed"),
			wantSummary:  "Error reading tag bindings for registry artifact",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := &tfe.Client{}
			expectListTagBindings(ctrl, client, testCase.artifactType, testCase.artifactID, testCase.tags, testCase.listErr)

			r := &resourceTFERegistryArtifactTag{config: ConfiguredClient{Client: client}}
			sch := artifactTagResourceSchema(t, r)
			model := artifactTagModel(testCase.artifactType, testCase.artifactID, testCase.artifactType+"/"+testCase.artifactID, map[string]string{"stale": "tag"})
			resp := &resource.ReadResponse{State: tfsdk.State{Schema: sch}}
			r.Read(ctx, resource.ReadRequest{State: mustArtifactTagState(t, sch, model)}, resp)

			if testCase.wantSummary != "" {
				assertDiagnosticSummary(t, resp.Diagnostics, testCase.wantSummary)
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("read diagnostics: %v", resp.Diagnostics)
			}
			if testCase.wantRemoved {
				if !resp.State.Raw.IsNull() {
					t.Fatal("expected resource state to be removed")
				}
				return
			}

			var got modelRegistryArtifactTag
			if diags := resp.State.Get(ctx, &got); diags.HasError() {
				t.Fatalf("state get: %v", diags)
			}
			if diff := diffRegistryArtifactTagMap(modelTagMap(got.Tags), testCase.tags); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestDataSourceRegistryArtifactTags_mockRead(t *testing.T) {
	testCases := map[string]struct {
		artifactType string
		artifactID   string
		tags         map[string]string
		listErr      error
		wantSummary  string
	}{
		"module": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-123",
			tags:         map[string]string{"env": "prod"},
		},
		"provider": {
			artifactType: ArtifactTypeRegistryProvider,
			artifactID:   "prov-123",
			tags:         map[string]string{"team": "platform"},
		},
		"component": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			tags:         map[string]string{"env": "dev"},
		},
		"not found": {
			artifactType: ArtifactTypeRegistryModule,
			artifactID:   "mod-missing",
			listErr:      tfe.ErrResourceNotFound,
			wantSummary:  "Registry Artifact Not Found",
		},
		"list error": {
			artifactType: ArtifactTypeRegistryComponent,
			artifactID:   "comp-123",
			listErr:      errors.New("list failed"),
			wantSummary:  "Error reading tags for registry artifact",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := &tfe.Client{}
			expectListTagBindings(ctrl, client, testCase.artifactType, testCase.artifactID, testCase.tags, testCase.listErr)

			d := &dataSourceTFERegistryArtifactTags{config: ConfiguredClient{Client: client}}
			sch := artifactTagDataSourceSchema(t, d)
			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: sch}}
			d.Read(ctx, datasource.ReadRequest{
				Config: mustArtifactTagConfig(t, sch, artifactTagModel(testCase.artifactType, testCase.artifactID, "", nil)),
			}, resp)

			if testCase.wantSummary != "" {
				assertDiagnosticSummary(t, resp.Diagnostics, testCase.wantSummary)
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("data source diagnostics: %v", resp.Diagnostics)
			}

			var got modelRegistryArtifactTagsData
			if diags := resp.State.Get(ctx, &got); diags.HasError() {
				t.Fatalf("state get: %v", diags)
			}
			wantID := testCase.artifactType + "/" + testCase.artifactID
			if got.ID.ValueString() != wantID {
				t.Fatalf("id = %q, want %q", got.ID.ValueString(), wantID)
			}
			if diff := diffRegistryArtifactTagMap(modelTagMap(got.Tags), testCase.tags); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestRegistryArtifactTag_mockUnsupportedType(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := &tfe.Client{
		RegistryModules:    tfemocks.NewMockRegistryModules(ctrl),
		RegistryProviders:  tfemocks.NewMockRegistryProviders(ctrl),
		RegistryComponents: tfemocks.NewMockRegistryComponents(ctrl),
	}
	r := &resourceTFERegistryArtifactTag{config: ConfiguredClient{Client: client}}

	err := r.updateTagBindings(context.Background(), "not-an-artifact", "id-1", []*tfe.TagBinding{{Key: "env", Value: "prod"}})
	if err == nil || !strings.Contains(err.Error(), "unsupported artifact type") {
		t.Fatalf("updateTagBindings error = %v", err)
	}

	_, err = helpers.ListRegistryArtifactTagBindings(context.Background(), client, "not-an-artifact", "id-1")
	if err == nil || !strings.Contains(err.Error(), "unsupported artifact type") {
		t.Fatalf("ListRegistryArtifactTagBindings error = %v", err)
	}
}

type artifactTagWrite struct {
	artifactType string
	artifactID   string
	tags         map[string]string
	readErr      error
	updateErr    error
	emptySlice   bool
}

func expectArtifactTagWrite(t *testing.T, ctrl *gomock.Controller, client *tfe.Client, write artifactTagWrite) {
	t.Helper()

	switch write.artifactType {
	case ArtifactTypeRegistryModule:
		mockModules := tfemocks.NewMockRegistryModules(ctrl)
		client.RegistryModules = mockModules
		read := mockModules.EXPECT().Read(gomock.Any(), tfe.RegistryModuleID{ID: write.artifactID})
		if write.readErr != nil {
			read.Return(nil, write.readErr)
			return
		}
		read.Return(&tfe.RegistryModule{
			ID:           write.artifactID,
			Name:         "vpc",
			Provider:     "aws",
			Namespace:    "hashicorp",
			RegistryName: tfe.PrivateRegistry,
			Organization: &tfe.Organization{Name: "hashicorp"},
		}, nil)
		mockModules.EXPECT().
			Update(gomock.Any(), tfe.RegistryModuleID{
				Name:         "vpc",
				RegistryName: tfe.PrivateRegistry,
				Namespace:    "hashicorp",
				Organization: "hashicorp",
				Provider:     "aws",
			}, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ tfe.RegistryModuleID, opts tfe.RegistryModuleUpdateOptions) (*tfe.RegistryModule, error) {
				assertWrittenTagBindings(t, opts.TagBindings, write)
				if write.updateErr != nil {
					return nil, write.updateErr
				}
				return &tfe.RegistryModule{}, nil
			})
	case ArtifactTypeRegistryProvider:
		mockProviders := tfemocks.NewMockRegistryProviders(ctrl)
		client.RegistryProviders = mockProviders
		read := mockProviders.EXPECT().Read(gomock.Any(), tfe.RegistryProviderID{ID: write.artifactID}, nil)
		if write.readErr != nil {
			read.Return(nil, write.readErr)
			return
		}
		read.Return(&tfe.RegistryProvider{
			ID:           write.artifactID,
			Name:         "aws",
			Namespace:    "hashicorp",
			RegistryName: tfe.PrivateRegistry,
			Organization: &tfe.Organization{Name: "hashicorp"},
		}, nil)
		mockProviders.EXPECT().
			Update(gomock.Any(), tfe.RegistryProviderID{
				Name:             "aws",
				RegistryName:     tfe.PrivateRegistry,
				Namespace:        "hashicorp",
				OrganizationName: "hashicorp",
			}, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ tfe.RegistryProviderID, opts *tfe.RegistryProviderUpdateOptions) (*tfe.RegistryProvider, error) {
				if opts == nil {
					t.Fatal("provider update options are nil")
				}
				assertWrittenTagBindings(t, opts.TagBindings, write)
				if write.updateErr != nil {
					return nil, write.updateErr
				}
				return &tfe.RegistryProvider{}, nil
			})
	case ArtifactTypeRegistryComponent:
		mockComponents := tfemocks.NewMockRegistryComponents(ctrl)
		client.RegistryComponents = mockComponents
		mockComponents.EXPECT().
			Update(gomock.Any(), write.artifactID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, opts *tfe.RegistryComponentUpdateOptions) (*tfe.RegistryComponent, error) {
				if opts == nil {
					t.Fatal("component update options are nil")
				}
				assertWrittenTagBindings(t, opts.TagBindings, write)
				if write.updateErr != nil {
					return nil, write.updateErr
				}
				return &tfe.RegistryComponent{}, nil
			})
	default:
		t.Fatalf("unsupported artifact type %q", write.artifactType)
	}
}

func expectListTagBindings(ctrl *gomock.Controller, client *tfe.Client, artifactType, artifactID string, tags map[string]string, listErr error) {
	bindings := make([]*tfe.TagBinding, 0, len(tags))
	for key, value := range tags {
		bindings = append(bindings, &tfe.TagBinding{Key: key, Value: value})
	}

	switch artifactType {
	case ArtifactTypeRegistryModule:
		mockModules := tfemocks.NewMockRegistryModules(ctrl)
		client.RegistryModules = mockModules
		call := mockModules.EXPECT().ListTagBindings(gomock.Any(), artifactID)
		if listErr != nil {
			call.Return(nil, listErr)
			return
		}
		call.Return(bindings, nil)
	case ArtifactTypeRegistryProvider:
		mockProviders := tfemocks.NewMockRegistryProviders(ctrl)
		client.RegistryProviders = mockProviders
		call := mockProviders.EXPECT().ListTagBindings(gomock.Any(), artifactID)
		if listErr != nil {
			call.Return(nil, listErr)
			return
		}
		call.Return(bindings, nil)
	case ArtifactTypeRegistryComponent:
		mockComponents := tfemocks.NewMockRegistryComponents(ctrl)
		client.RegistryComponents = mockComponents
		call := mockComponents.EXPECT().ListTagBindings(gomock.Any(), artifactID)
		if listErr != nil {
			call.Return(nil, listErr)
			return
		}
		call.Return(bindings, nil)
	default:
		panic("unsupported artifact type " + artifactType)
	}
}

func assertWrittenTagBindings(t *testing.T, got []*tfe.TagBinding, write artifactTagWrite) {
	t.Helper()
	if write.emptySlice && got == nil {
		t.Fatal("expected empty tag binding slice, got nil")
	}
	if diff := diffRegistryArtifactTagBindings(got, write.tags); diff != "" {
		t.Fatal(diff)
	}
}

func assertDiagnosticSummary(t *testing.T, diags diag.Diagnostics, want string) {
	t.Helper()
	errs := diags.Errors()
	if len(errs) == 0 {
		t.Fatalf("expected diagnostic %q", want)
	}
	if errs[0].Summary() != want {
		t.Fatalf("summary = %q, want %q", errs[0].Summary(), want)
	}
}

func artifactTagModel(artifactType, artifactID, id string, tags map[string]string) modelRegistryArtifactTag {
	modelTags := make([]modelTag, 0, len(tags))
	for key, value := range tags {
		modelTags = append(modelTags, modelTag{
			Key:   types.StringValue(key),
			Value: types.StringValue(value),
		})
	}
	idValue := types.StringNull()
	if id != "" {
		idValue = types.StringValue(id)
	}
	if modelTags == nil {
		modelTags = []modelTag{}
	}
	return modelRegistryArtifactTag{
		ID:   idValue,
		Tags: modelTags,
		Artifact: modelArtifact{
			Type: types.StringValue(artifactType),
			ID:   types.StringValue(artifactID),
		},
	}
}

func modelTagMap(tags []modelTag) map[string]string {
	out := make(map[string]string, len(tags))
	for _, tag := range tags {
		out[tag.Key.ValueString()] = tag.Value.ValueString()
	}
	return out
}

func artifactTagResourceSchema(t *testing.T, r *resourceTFERegistryArtifactTag) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func artifactTagDataSourceSchema(t *testing.T, d *dataSourceTFERegistryArtifactTags) datasourceschema.Schema {
	t.Helper()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func mustArtifactTagPlan(t *testing.T, sch schema.Schema, model modelRegistryArtifactTag) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: sch}
	if diags := plan.Set(ctx, &model); diags.HasError() {
		t.Fatalf("plan set: %v", diags)
	}
	return plan
}

func mustArtifactTagState(t *testing.T, sch schema.Schema, model modelRegistryArtifactTag) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: sch}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("state set: %v", diags)
	}
	return state
}

func mustArtifactTagConfig(t *testing.T, sch datasourceschema.Schema, model modelRegistryArtifactTag) tfsdk.Config {
	t.Helper()
	state := tfsdk.State{Schema: sch}
	data := modelRegistryArtifactTagsData{
		ID:       model.ID,
		Tags:     model.Tags,
		Artifact: model.Artifact,
	}
	if diags := state.Set(ctx, &data); diags.HasError() {
		t.Fatalf("config set: %v", diags)
	}
	return tfsdk.Config{Raw: state.Raw, Schema: sch}
}
