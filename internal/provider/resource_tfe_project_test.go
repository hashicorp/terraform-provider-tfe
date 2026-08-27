// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// notFoundProjectHandler returns a handler that 404s a single project ID, and everything else,
// for the "removed project" Read unit tests below.
func notFoundProjectHandler(projectID string) http.Handler {
	mux := http.NewServeMux()
	notFound := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
	}
	mux.HandleFunc("/api/v2/projects/"+projectID, notFound)
	mux.HandleFunc("/", notFound)
	return mux
}

func TestAccTFEProject_basic(t *testing.T) {
	var project models.Projectsable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFEProject_invalidName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProject_invalidNameChar(rInt),
				ExpectError: regexp.MustCompile(`can only include letters, numbers, spaces, -, and _.`),
			},
			{
				Config:      testAccTFEProject_invalidNameLen(rInt),
				ExpectError: regexp.MustCompile(`string length must be between 3 and 40`),
			},
		},
	})
}

func TestAccTFEProject_update(t *testing.T) {
	var project models.Projectsable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
				),
			},
			{
				Config: testAccTFEProject_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributesUpdated(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "project updated"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description updated"),
				),
			},
			{
				Config: testAccTFEProject_updateRemoveBindings(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributesUpdated(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "project updated"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description updated"),
				),
			},
		},
	})
}

func TestAccTFEProject_ignoreAdditionalTags(t *testing.T) {
	var project models.Projectsable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_ignoreAdditionalTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
				),
			},
			{
				Config: testAccTFEProject_ignoreAdditionalTags(rInt),
				PreConfig: func() {
					organization := fmt.Sprintf("tst-terraform-%d", rInt)
					projects, err := testAccConfiguredClient.Client.Projects.List(ctx, organization, &tfe.ProjectListOptions{Name: "projecttest"})
					if err != nil {
						t.Fatalf("failed reading projecttest: %v", err)
					}
					if len(projects.Items) == 0 {
						t.Fatalf("expected to find projecttest, for %s", organization)
					}

					_, err = testAccConfiguredClient.Client.Projects.AddTagBindings(ctx, projects.Items[0].ID, tfe.ProjectAddTagBindingsOptions{
						TagBindings: []*tfe.TagBinding{{
							Key:   "additional",
							Value: "tag",
						}},
					})
					if err != nil {
						t.Fatalf("failed adding tag binding via API call: %v", err)
					}
				},
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccTFEProject_tagBindings(t *testing.T) {
	var project models.Projectsable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basicTagBindings(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyB", "valueB"),
				),
			},
			{
				Config: testAccTFEProject_basicTagBindingsAddOne(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "3"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyA", "valueA"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyB", "valueB"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.keyC", "valueC"),
				),
			},
			{
				Config: testAccTFEProject_basicTagBindingsRemoveAll(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "name", "projecttest"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "description", "project description"),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "organization", fmt.Sprintf("tst-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccTFEProject_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	var project models.Projectsable
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				Check: testAccCheckTFEProjectExists(
					"tfe_project.foobar", &project),
			},

			{
				ResourceName:      "tfe_project.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: "tfe_project.foobar",
				ImportState:  true,
				// project is populated by the Check callback in the first step's TestStep,
				// which runs at test-execution time; this struct literal is built eagerly
				// before that, so project is always unset here (matching prior behavior) and
				// this always imports by the resource's own ID rather than a stale one.
				ImportStateId:     "",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEProject_importByIdentity(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basic(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_project.foobar", map[string]knownvalue.Check{
						"id":       knownvalue.NotNull(),
						"hostname": knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},
			{
				ResourceName:    "tfe_project.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestResourceTFEProjectRead_RemovedProjectBackfillsIdentity(t *testing.T) {
	ctx := context.Background()
	clientV2 := testTfeClientV2(t, notFoundProjectHandler("prj-123"))

	r := &resourceTFEProject{config: ConfiguredClient{ClientV2: clientV2}}

	readResp := runRemovedProjectRead(t, ctx, r, modelTFEProject{
		ID:                          types.StringValue("prj-123"),
		Name:                        types.StringValue("projecttest"),
		Description:                 types.StringValue("project description"),
		Organization:                types.StringValue("example-org"),
		AutoDestroyActivityDuration: types.StringNull(),
		Tags:                        types.MapNull(types.StringType),
		IgnoreAdditionalTags:        types.BoolValue(false),
	})

	assertRemovedProjectRead(t, ctx, readResp, modelProjectIdentity{
		ID:       types.StringValue("prj-123"),
		Hostname: types.StringValue(clientV2.BaseURL().Host),
	})
}

func TestResourceTFEProjectRead_RemovedProjectPreservesExistingIdentity(t *testing.T) {
	ctx := context.Background()
	clientV2 := testTfeClientV2(t, notFoundProjectHandler("prj-123"))

	r := &resourceTFEProject{config: ConfiguredClient{ClientV2: clientV2}}
	existingIdentity := &modelProjectIdentity{
		ID:       types.StringValue("prj-existing"),
		Hostname: types.StringValue("preserve.example.com"),
	}

	readResp := runRemovedProjectRead(t, ctx, r, modelTFEProject{
		ID:                          types.StringValue("prj-123"),
		Name:                        types.StringValue("projecttest"),
		Description:                 types.StringValue("project description"),
		Organization:                types.StringValue("example-org"),
		AutoDestroyActivityDuration: types.StringNull(),
		Tags:                        types.MapNull(types.StringType),
		IgnoreAdditionalTags:        types.BoolValue(false),
	}, existingIdentity)

	assertRemovedProjectRead(t, ctx, readResp, *existingIdentity)
}

func runRemovedProjectRead(t *testing.T, ctx context.Context, r *resourceTFEProject, stateData modelTFEProject, existingIdentity ...*modelProjectIdentity) fwresource.ReadResponse {
	t.Helper()

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &stateData); diags.HasError() {
		t.Fatalf("unexpected state set diagnostics: %v", diags)
	}

	identitySchemaResp := &fwresource.IdentitySchemaResponse{}
	r.IdentitySchema(ctx, fwresource.IdentitySchemaRequest{}, identitySchemaResp)
	nullIdentity := tftypes.NewValue(identitySchemaResp.IdentitySchema.Type().TerraformType(ctx), nil)

	requestIdentity := &tfsdk.ResourceIdentity{Schema: identitySchemaResp.IdentitySchema, Raw: nullIdentity.Copy()}
	responseIdentity := &tfsdk.ResourceIdentity{Schema: identitySchemaResp.IdentitySchema, Raw: nullIdentity.Copy()}

	if len(existingIdentity) > 0 && existingIdentity[0] != nil {
		if diags := requestIdentity.Set(ctx, existingIdentity[0]); diags.HasError() {
			t.Fatalf("unexpected request identity diagnostics: %v", diags)
		}
		if diags := responseIdentity.Set(ctx, existingIdentity[0]); diags.HasError() {
			t.Fatalf("unexpected response identity diagnostics: %v", diags)
		}
	}

	readResp := fwresource.ReadResponse{
		State:    tfsdk.State{Schema: schemaResp.Schema, Raw: state.Raw.Copy()},
		Identity: responseIdentity,
	}

	r.Read(ctx, fwresource.ReadRequest{
		State:    tfsdk.State{Schema: schemaResp.Schema, Raw: state.Raw.Copy()},
		Identity: requestIdentity,
	}, &readResp)

	return readResp
}

func assertRemovedProjectRead(t *testing.T, ctx context.Context, readResp fwresource.ReadResponse, expectedIdentity modelProjectIdentity) {
	t.Helper()

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected read diagnostics: %v", readResp.Diagnostics)
	}

	if !readResp.State.Raw.IsFullyNull() {
		t.Fatalf("expected resource to be removed from state, got %s", readResp.State.Raw.String())
	}

	if readResp.Identity == nil || readResp.Identity.Raw.IsFullyNull() {
		t.Fatal("expected project identity to be preserved for removed resource")
	}

	var gotIdentity modelProjectIdentity
	if diags := readResp.Identity.Get(ctx, &gotIdentity); diags.HasError() {
		t.Fatalf("unexpected identity diagnostics: %v", diags)
	}

	if gotIdentity.ID.ValueString() != expectedIdentity.ID.ValueString() {
		t.Fatalf("expected identity id %q, got %q", expectedIdentity.ID.ValueString(), gotIdentity.ID.ValueString())
	}

	if gotIdentity.Hostname.ValueString() != expectedIdentity.Hostname.ValueString() {
		t.Fatalf("expected hostname %q, got %q", expectedIdentity.Hostname.ValueString(), gotIdentity.Hostname.ValueString())
	}
}

func TestAccTFEProject_withAutoDestroy(t *testing.T) {
	var project models.Projectsable
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProject_basicWithAutoDestroy(rInt, "3d"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckResourceAttr(
						"tfe_project.foobar", "auto_destroy_activity_duration", "3d"),
				),
			},
			{
				Config:      testAccTFEProject_basicWithAutoDestroy(rInt, "10m"),
				ExpectError: regexp.MustCompile("must be 1-4 digits followed by"),
			},
			{
				Config: testAccTFEProject_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProjectExists(
						"tfe_project.foobar", &project),
					testAccCheckTFEProjectAttributes(&project),
					resource.TestCheckNoResourceAttr("tfe_project.foobar", "auto_destroy_activity_duration"),
				),
			},
		},
	})
}

func testAccTFEProject_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project updated"
  description = "project description updated"
}`, rInt)
}

func testAccTFEProject_updateRemoveBindings(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "project updated"
  description = "project description updated"
}`, rInt)
}

func testAccTFEProject_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
}`, rInt)
}

func testAccTFEProject_basicTagBindings(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
  }
}`, rInt)
}

func testAccTFEProject_basicTagBindingsAddOne(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
	  keyC = "valueC"
  }
}`, rInt)
}

func testAccTFEProject_basicTagBindingsRemoveAll(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {}
}`, rInt)
}

func testAccTFEProject_ignoreAdditionalTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  description = "project description"
  tags = {
	  keyA = "valueA"
	  keyB = "valueB"
  }
  ignore_additional_tags = true
}`, rInt)
}

func testAccTFEProject_invalidNameChar(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "invalidchar#"
}`, rInt)
}
func testAccTFEProject_invalidNameLen(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "aa"
}`, rInt)
}

func testAccTFEProject_basicWithAutoDestroy(rInt int, duration string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_project" "foobar" {
  organization = tfe_organization.foobar.name
  name = "projecttest"
  auto_destroy_activity_duration = "%s"
}`, rInt, duration)
}

func testAccCheckTFEProjectDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_project" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.ClientV2.API.Projects().ByProject_id(rs.Primary.ID).Get(ctx, nil)
		if err == nil {
			return fmt.Errorf("Project %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTFEProjectExists(n string, project *models.Projectsable) resource.TestCheckFunc { //nolint:gocritic // project is an output param the caller reads after Check runs; Projectsable must stay addressable to be settable
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		projEnvelope, err := testAccConfiguredClient.ClientV2.API.Projects().ByProject_id(rs.Primary.ID).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("unable to read project with ID %s", rs.Primary.ID)
		}
		if projEnvelope == nil || projEnvelope.GetData() == nil {
			return fmt.Errorf("unable to read project with ID %s", rs.Primary.ID)
		}

		*project = projEnvelope.GetData()

		return nil
	}
}

// projectName reads the name off a *models.Projectsable output param, dereferencing it only when
// the closure runs (by which time testAccCheckTFEProjectExists has populated it).
func projectName(project *models.Projectsable) string { //nolint:gocritic // project is populated by the paired Exists check at test-execution time; must stay a pointer so this reads that value, not a stale copy captured at construction time
	p := *project
	if p == nil || p.GetAttributes() == nil {
		return ""
	}
	return valueOrZero(p.GetAttributes().GetName())
}

func testAccCheckTFEProjectAttributes(
	project *models.Projectsable) resource.TestCheckFunc { //nolint:gocritic // project is populated by the paired Exists check at test-execution time; must stay a pointer so this reads that value, not a stale copy captured at construction time
	return func(s *terraform.State) error {
		if name := projectName(project); name != "projecttest" {
			return fmt.Errorf("Bad name: %s", name)
		}

		return nil
	}
}

func testAccCheckTFEProjectAttributesUpdated(
	project *models.Projectsable) resource.TestCheckFunc { //nolint:gocritic // project is populated by the paired Exists check at test-execution time; must stay a pointer so this reads that value, not a stale copy captured at construction time
	return func(s *terraform.State) error {
		if name := projectName(project); name != "project updated" {
			return fmt.Errorf("Bad name: %s", name)
		}

		return nil
	}
}
