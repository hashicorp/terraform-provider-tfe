// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type notFoundStacks struct{}

func (notFoundStacks) List(_ context.Context, _ string, _ *tfe.StackListOptions) (*tfe.StackList, error) {
	return nil, nil
}

func (notFoundStacks) Read(_ context.Context, _ string) (*tfe.Stack, error) {
	return nil, tfe.ErrResourceNotFound
}

func (notFoundStacks) Create(_ context.Context, _ tfe.StackCreateOptions) (*tfe.Stack, error) {
	return nil, nil
}

func (notFoundStacks) Update(_ context.Context, _ string, _ tfe.StackUpdateOptions) (*tfe.Stack, error) {
	return nil, nil
}

func (notFoundStacks) Delete(_ context.Context, _ string) error {
	return nil
}

func (notFoundStacks) ForceDelete(_ context.Context, _ string) error {
	return nil
}

func (notFoundStacks) FetchLatestFromVcs(_ context.Context, _ string) (*tfe.Stack, error) {
	return nil, nil
}

func TestAccTFEStackResource_basic(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	speculativeEnabledFalse := false

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", true, speculativeEnabledFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "project_id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "agent_pool_id"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "name", "example-stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "description", "Just an ordinary stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "vcs_repo.identifier", "arunatibm/pet-nulls-stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "creation_source", "migration-api"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "vcs_repo.oauth_token_id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "speculative_enabled"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "updated_at"),
				),
			},
			{
				ResourceName:            "tfe_stack.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"migration"},
			},
		},
	})
}

func TestAccTFEStackResource_omitSpeculativeEnabled(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_stack.foobar", "speculative_enabled", "false"),
				),
			},
			{
				ResourceName:            "tfe_stack.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"migration"},
			},
		},
	})
}

func TestAccTFEStackResource_updateSpeculativeEnabled(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)
	speculativeEnabledFalse := false
	speculativeEnabledTrue := true

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", false, speculativeEnabledFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_stack.foobar", "speculative_enabled", "false"),
				),
			},
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", true, speculativeEnabledFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_stack.foobar", "speculative_enabled", "false"),
				),
			},
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", true, speculativeEnabledTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_stack.foobar", "speculative_enabled", "true"),
				),
			},
		},
	})
}

func TestAccTFEStackResource_importByIdentity(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfig(orgName, envGithubToken, "arunatibm/pet-nulls-stack", true, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_stack.foobar", map[string]knownvalue.Check{
						"id":       knownvalue.NotNull(),
						"hostname": knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},
			{
				ResourceName:    "tfe_stack.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func testAccTFEStackResourceConfig(orgName, ghToken, ghRepoIdentifier string, includeSpeculativeEnabled bool, speculativeEnabled bool) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
  stacks_enabled = true
}

resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-example"
  organization          = tfe_organization.foobar.name
}

resource "tfe_project" "example" {
	name         = "example"
	organization = tfe_organization.foobar.name
}

resource "tfe_oauth_client" "foobar" {
  organization     = tfe_organization.foobar.name
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "%s"
  service_provider = "github"
}

resource "tfe_stack" "foobar" {
	name        = "example-stack"
	description = "Just an ordinary stack"
  project_id  = tfe_project.example.id
  agent_pool_id = tfe_agent_pool.foobar.id
	migration = true
	vcs_repo {
    identifier         = "%s"
    oauth_token_id     = tfe_oauth_client.foobar.oauth_token_id
  }
`, orgName, ghToken, ghRepoIdentifier))
	if includeSpeculativeEnabled {
		builder.WriteString(fmt.Sprintf(`	speculative_enabled = "%t"`, speculativeEnabled))
	}
	builder.WriteString("\n}")
	return builder.String()
}

func TestAccTFEStackResource_withAgentPool(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfigWithAgentPool(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "project_id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "agent_pool_id"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "name", "example-stack"),
					resource.TestCheckResourceAttr("tfe_stack.foobar", "description", "Just an ordinary stack"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar", "updated_at"),
				),
			},
			{
				ResourceName:      "tfe_stack.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEStackResourceConfigWithAgentPool(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_agent_pool" "foobar" {
  name                  = "agent-pool-test-example"
  organization          = tfe_organization.foobar.name
}

resource "tfe_project" "example" {
	name         = "example"
	organization = tfe_organization.foobar.name
}

resource "tfe_stack" "foobar" {
	name        = "example-stack"
	description = "Just an ordinary stack"
    project_id  = tfe_project.example.id
    agent_pool_id = tfe_agent_pool.foobar.id
}
`, orgName)
}

func TestResourceTFEStackRead_RemovedStackBackfillsIdentity(t *testing.T) {
	ctx := context.Background()
	client := testTfeClient(t, testClientOptions{})
	client.Stacks = notFoundStacks{}

	r := &resourceTFEStack{config: ConfiguredClient{Client: client}}

	readResp := runRemovedStackRead(t, ctx, r, modelTFEStack{
		ID:                 types.StringValue("stack-123"),
		ProjectID:          types.StringValue("prj-123"),
		AgentPoolID:        types.StringNull(),
		Name:               types.StringValue("test-stack"),
		Migration:          types.BoolValue(false),
		SpeculativeEnabled: types.BoolValue(false),
		CreationSource:     types.StringNull(),
		Description:        types.StringValue(""),
		VCSRepo:            nil,
		CreatedAt:          types.StringValue("2026-01-01T00:00:00Z"),
		UpdatedAt:          types.StringValue("2026-01-01T00:00:00Z"),
	})

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected read diagnostics: %v", readResp.Diagnostics)
	}

	if !readResp.State.Raw.IsFullyNull() {
		t.Fatalf("expected stack to be removed from state, got %s", readResp.State.Raw.String())
	}

	if readResp.Identity == nil || readResp.Identity.Raw.IsFullyNull() {
		t.Fatal("expected stack identity to be preserved for removed resource")
	}

	var gotIdentity modelTFEStackIdentity
	if diags := readResp.Identity.Get(ctx, &gotIdentity); diags.HasError() {
		t.Fatalf("unexpected identity diagnostics: %v", diags)
	}

	if gotIdentity.ID.ValueString() != "stack-123" {
		t.Fatalf("expected identity id %q, got %q", "stack-123", gotIdentity.ID.ValueString())
	}

	if gotIdentity.Hostname.ValueString() != client.BaseURL().Host {
		t.Fatalf("expected hostname %q, got %q", client.BaseURL().Host, gotIdentity.Hostname.ValueString())
	}
}

func runRemovedStackRead(t *testing.T, ctx context.Context, r *resourceTFEStack, stateData modelTFEStack) fwresource.ReadResponse {
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

	requestIdentity := &tfsdk.ResourceIdentity{
		Schema: identitySchemaResp.IdentitySchema,
		Raw:    nullIdentity.Copy(),
	}
	responseIdentity := &tfsdk.ResourceIdentity{
		Schema: identitySchemaResp.IdentitySchema,
		Raw:    nullIdentity.Copy(),
	}

	readResp := fwresource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    state.Raw.Copy(),
		},
		Identity: responseIdentity,
	}

	r.Read(ctx, fwresource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    state.Raw.Copy(),
		},
		Identity: requestIdentity,
	}, &readResp)

	return readResp
}

func TestAccTFEStackResource_noVCSRepo(t *testing.T) {
	skipUnlessBeta(t)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStackResourceConfigNoVCSRepo(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_stack.foobar2", "id"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar2", "project_id"),
					resource.TestCheckResourceAttr("tfe_stack.foobar2", "name", "example-stack-no-vcs"),
					resource.TestCheckResourceAttr("tfe_stack.foobar2", "description", "Stack without VCS repo"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar2", "created_at"),
					resource.TestCheckResourceAttrSet("tfe_stack.foobar2", "updated_at"),
				),
			},
			{
				ResourceName:      "tfe_stack.foobar2",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTFEStackResourceConfigNoVCSRepo(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "%s"
  email = "admin@tfe.local"
}

resource "tfe_project" "example" {
	name         = "example"
	organization = tfe_organization.foobar.name
}

resource "tfe_stack" "foobar2" {
	name        = "example-stack-no-vcs"
	description = "Stack without VCS repo"
	project_id  = tfe_project.example.id
}
`, orgName)
}
