// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe/v2/api/models"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestTaskConfigValidateConfig(t *testing.T) {
	cases := map[string]struct {
		global      tftypes.Value
		projectIDs  tftypes.Value
		expectError bool
	}{
		"global with projects":       {tftypes.NewValue(tftypes.Bool, true), taskConfigTestSet("prj-1"), true},
		"global with unknown":        {tftypes.NewValue(tftypes.Bool, true), tftypes.NewValue(taskConfigTestSetType, tftypes.UnknownValue), true},
		"global with empty set":      {tftypes.NewValue(tftypes.Bool, true), taskConfigTestSet(), false},
		"global without projects":    {tftypes.NewValue(tftypes.Bool, true), tftypes.NewValue(taskConfigTestSetType, nil), false},
		"non-global with projects":   {tftypes.NewValue(tftypes.Bool, false), taskConfigTestSet("prj-1"), false},
		"unset global with projects": {tftypes.NewValue(tftypes.Bool, nil), taskConfigTestSet("prj-1"), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &resourceTaskConfig{}
			schemaResp := &fwresource.SchemaResponse{}
			r.Schema(context.Background(), fwresource.SchemaRequest{}, schemaResp)

			raw := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), map[string]tftypes.Value{
				"id":                tftypes.NewValue(tftypes.String, nil),
				"organization":      tftypes.NewValue(tftypes.String, nil),
				"owner_id":          tftypes.NewValue(tftypes.String, nil),
				"task_id":           tftypes.NewValue(tftypes.String, "task-123"),
				"global":            tc.global,
				"enforcement_level": tftypes.NewValue(tftypes.String, "advisory"),
				"stages":            taskConfigTestSet("pre_plan"),
				"project_ids":       tc.projectIDs,
			})

			resp := &fwresource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), fwresource.ValidateConfigRequest{
				Config: tfsdk.Config{Raw: raw, Schema: schemaResp.Schema},
			}, resp)

			if got := resp.Diagnostics.HasError(); got != tc.expectError {
				t.Errorf("expected error=%t, got diagnostics: %v", tc.expectError, resp.Diagnostics)
			}
		})
	}
}

var taskConfigTestSetType = tftypes.Set{ElementType: tftypes.String}

func taskConfigTestSet(values ...string) tftypes.Value {
	elems := make([]tftypes.Value, len(values))
	for i, v := range values {
		elems[i] = tftypes.NewValue(tftypes.String, v)
	}
	return tftypes.NewValue(taskConfigTestSetType, elems)
}

func TestNewTaskConfigCreateEnvelope(t *testing.T) {
	envelope, err := newTaskConfigCreateEnvelope("task-123", "example-org", taskConfigInput{
		global:           true,
		stages:           []string{"pre_plan", "post_apply"},
		enforcementLevel: "mandatory",
		projectIDs:       []string{"prj-1", "prj-2"},
	})
	if err != nil {
		t.Fatal(err)
	}

	data := envelope.GetData()
	if data == nil || data.GetAttributes() == nil {
		t.Fatal("expected task config data and attributes")
	}
	if data.GetId() != nil {
		t.Errorf("expected no ID on create, got %q", *data.GetId())
	}

	attributes := data.GetAttributes()
	if !valueOrZero(attributes.GetGlobal()) {
		t.Error("expected global to be true")
	}
	if got := attributes.GetEnforcementLevel(); got == nil || got.String() != "mandatory" {
		t.Errorf("expected enforcement level mandatory, got %v", got)
	}
	if got := attributes.GetAllowedStages(); len(got) != 2 || got[0].String() != "pre_plan" || got[1].String() != "post_apply" {
		t.Errorf("unexpected allowed stages: %v", got)
	}

	relationships := data.GetRelationships()
	if relationships == nil {
		t.Fatal("expected relationships")
	}
	if relationships.GetTask() == nil || relationships.GetTask().GetData() == nil || valueOrZero(relationships.GetTask().GetData().GetId()) != "task-123" {
		t.Errorf("expected task relationship task-123, got %#v", relationships.GetTask())
	}
	if relationships.GetOwner() == nil || relationships.GetOwner().GetData() == nil || valueOrZero(relationships.GetOwner().GetData().GetId()) != "example-org" {
		t.Errorf("expected owner relationship example-org, got %#v", relationships.GetOwner())
	}
	if relationships.GetProjects() == nil || len(relationships.GetProjects().GetData()) != 2 {
		t.Fatalf("expected 2 project relationships, got %#v", relationships.GetProjects())
	}
	if got := valueOrZero(relationships.GetProjects().GetData()[1].GetId()); got != "prj-2" {
		t.Errorf("expected second project prj-2, got %q", got)
	}
}

func TestNewTaskConfigUpdateEnvelope(t *testing.T) {
	envelope, err := newTaskConfigUpdateEnvelope("tc-123", taskConfigInput{
		stages:           []string{"post_plan"},
		enforcementLevel: "advisory",
	})
	if err != nil {
		t.Fatal(err)
	}

	data := envelope.GetData()
	if got := valueOrZero(data.GetId()); got != "tc-123" {
		t.Errorf("expected ID tc-123, got %q", got)
	}
	if data.GetAttributes() == nil || data.GetAttributes().GetGlobal() == nil || *data.GetAttributes().GetGlobal() {
		t.Error("expected global to be explicitly false")
	}
	relationships := data.GetRelationships()
	if relationships.GetTask() != nil || relationships.GetOwner() != nil {
		t.Error("expected task and owner to be omitted on update")
	}
	if relationships.GetProjects() == nil || relationships.GetProjects().GetData() == nil || len(relationships.GetProjects().GetData()) != 0 {
		t.Errorf("expected an empty projects relationship so projects can be cleared, got %#v", relationships.GetProjects())
	}
}

func TestNewTaskConfigAttributesInvalid(t *testing.T) {
	if _, err := newTaskConfigAttributes(taskConfigInput{stages: []string{"nope"}, enforcementLevel: "advisory"}); err == nil {
		t.Error("expected an error for an invalid stage")
	}
	if _, err := newTaskConfigAttributes(taskConfigInput{stages: []string{"pre_plan"}, enforcementLevel: "nope"}); err == nil {
		t.Error("expected an error for an invalid enforcement level")
	}
}

func TestTaskConfigCreateRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/example-org/task-configs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var payload struct {
			Data struct {
				Type       string `json:"type"`
				Attributes struct {
					Global           *bool    `json:"global"`
					AllowedStages    []string `json:"allowed-stages"`
					EnforcementLevel string   `json:"enforcement-level"`
				} `json:"attributes"`
				Relationships map[string]struct {
					Data json.RawMessage `json:"data"`
				} `json:"relationships"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Data.Type != "task-configs" {
			t.Errorf("expected type task-configs, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Global == nil || *payload.Data.Attributes.Global {
			t.Error("expected global to be false")
		}
		if got := payload.Data.Attributes.AllowedStages; len(got) != 1 || got[0] != "post_plan" {
			t.Errorf("unexpected allowed stages %v", got)
		}
		if got := payload.Data.Attributes.EnforcementLevel; got != "advisory" {
			t.Errorf("unexpected enforcement level %q", got)
		}
		for _, name := range []string{"task", "owner", "projects"} {
			if _, ok := payload.Data.Relationships[name]; !ok {
				t.Errorf("expected %q relationship in payload", name)
			}
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"data":{"id":"tc-123","type":"task-configs"}}`)
	})

	envelope, err := newTaskConfigCreateEnvelope("task-123", "example-org", taskConfigInput{
		stages:           []string{"post_plan"},
		enforcementLevel: "advisory",
		projectIDs:       []string{"prj-1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	client := testTfeClientV2(t, mux)
	created, err := client.API.Organizations().ByOrganization_name("example-org").TaskConfigs().Post(ctx, envelope, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := valueOrZero(created.GetData().GetId()); got != "tc-123" {
		t.Errorf("expected tc-123, got %q", got)
	}
}

func TestModelFromTFETaskConfig(t *testing.T) {
	attributes := models.NewTaskConfigs_attributes()
	attributes.SetGlobal(ptr(false))
	attributes.SetEnforcementLevel(ptr(models.ADVISORY_TASKCONFIGS_ATTRIBUTES_ENFORCEMENTLEVEL))
	attributes.SetAllowedStages([]models.TaskConfigs_attributes_allowedStages{
		models.PRE_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
		models.POST_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
	})

	envelope, err := newTaskConfigCreateEnvelope("task-123", "example-org", taskConfigInput{
		stages:           []string{"pre_plan"},
		enforcementLevel: "advisory",
		projectIDs:       []string{"prj-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	config := envelope.GetData()
	config.SetId(ptr("tc-123"))
	config.SetAttributes(attributes)

	result, diags := modelFromTFETaskConfig(config, "fallback-org")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result.ID.ValueString() != "tc-123" || result.TaskID.ValueString() != "task-123" {
		t.Errorf("unexpected identity id=%q task_id=%q", result.ID.ValueString(), result.TaskID.ValueString())
	}
	if result.Organization.ValueString() != "example-org" {
		t.Errorf("expected organization from owner relationship, got %q", result.Organization.ValueString())
	}
	if result.OwnerID.ValueString() != "example-org" {
		t.Errorf("expected owner_id from owner relationship, got %q", result.OwnerID.ValueString())
	}
	if result.Global.ValueBool() {
		t.Error("expected global to be false")
	}
	if result.EnforcementLevel.ValueString() != "advisory" {
		t.Errorf("unexpected enforcement level %q", result.EnforcementLevel.ValueString())
	}
	if got := result.Stages.Elements(); len(got) != 2 {
		t.Errorf("unexpected stages: %v", got)
	}
	if got := result.ProjectIDs.Elements(); len(got) != 1 || got[0].String() != `"prj-1"` {
		t.Errorf("unexpected project IDs: %v", got)
	}
}

func TestModelFromTFETaskConfigFallbackOrganization(t *testing.T) {
	config := models.NewTaskConfigs()
	config.SetId(ptr("tc-123"))

	result, diags := modelFromTFETaskConfig(config, "fallback-org")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result.Organization.ValueString() != "fallback-org" {
		t.Errorf("expected fallback organization, got %q", result.Organization.ValueString())
	}
	if result.OwnerID.ValueString() != "fallback-org" {
		t.Errorf("expected fallback owner_id, got %q", result.OwnerID.ValueString())
	}
	if result.ProjectIDs.IsNull() || len(result.ProjectIDs.Elements()) != 0 {
		t.Errorf("expected an empty, non-null project_ids set, got %v", result.ProjectIDs)
	}
}

func TestAccTFETaskConfig_validateSchemaAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFETaskConfig_parameters("", `["pre_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute enforcement_level value must be one of: \[.*\]`),
			},
			{
				Config:      testAccTFETaskConfig_parameters("mandatory", `[]`),
				ExpectError: regexp.MustCompile(`Attribute stages set must contain at least 1 elements.*`),
			},
			{
				Config:      testAccTFETaskConfig_parameters("mandatory", `["pre_plan","BADWOLF"]`),
				ExpectError: regexp.MustCompile(`Attribute stages\[.*\] value must be.*`),
			},
			{
				Config:      testAccTFETaskConfig_globalWithProjects(),
				ExpectError: regexp.MustCompile(`cannot be set when .global. is true`),
			},
		},
	})
}

func TestAccTFETaskConfig_basic(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETaskConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETaskConfig_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_task_config.sut", "id"),
					resource.TestCheckResourceAttrPair("tfe_task_config.sut", "task_id", "tfe_organization_run_task.foobar", "id"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "organization", org.Name),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "owner_id", org.Name),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "global", "true"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "stages.#", "1"),
					resource.TestCheckTypeSetElemAttr("tfe_task_config.sut", "stages.*", "post_plan"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "project_ids.#", "0"),
				),
			},
			{
				Config: testAccTFETaskConfig_update(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_task_config.sut", "global", "false"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "stages.#", "2"),
					resource.TestCheckTypeSetElemAttr("tfe_task_config.sut", "stages.*", "pre_plan"),
					resource.TestCheckTypeSetElemAttr("tfe_task_config.sut", "stages.*", "post_plan"),
					resource.TestCheckResourceAttr("tfe_task_config.sut", "project_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("tfe_task_config.sut", "project_ids.*", "tfe_project.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFETaskConfig_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFETaskConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFETaskConfig_update(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
			},
			{
				ResourceName:      "tfe_task_config.sut",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFETaskConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_task_config" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.ClientV2.API.TaskConfigs().ByExternal_id(rs.Primary.ID).Get(ctx, nil)
		if err == nil {
			return fmt.Errorf("Task config %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFETaskConfig_basic(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
  organization = "%s"
  url          = "%s"
  name         = "foobar-task-%d"
  enabled      = true
  hmac_key     = "%s"
}

resource "tfe_task_config" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  global            = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFETaskConfig_update(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
  organization = "%[1]s"
  url          = "%[2]s"
  name         = "foobar-task-%[3]d"
  enabled      = true
  hmac_key     = "%[4]s"
}

resource "tfe_project" "foobar" {
  organization = "%[1]s"
  name         = "foobar-project-%[3]d"
}

resource "tfe_task_config" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  global            = false
  enforcement_level = "advisory"
  stages            = ["pre_plan", "post_plan"]
  project_ids       = [tfe_project.foobar.id]
}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFETaskConfig_globalWithProjects() string {
	return `
resource "tfe_task_config" "sut" {
  task_id           = "task-123"
  global            = true
  enforcement_level = "advisory"
  stages            = ["pre_plan"]
  project_ids       = ["prj-123"]
}
`
}

func testAccTFETaskConfig_parameters(enforceLevel, stages string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
  organization = "foo"
  url          = "http://somewhere.local"
  name         = "task_name"
  enabled      = true
  hmac_key     = "something"
}

resource "tfe_task_config" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  enforcement_level = "%s"
  stages            = %s
}
`, enforceLevel, stages)
}
