// Copyright IBM Corp. 2018, 2025
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

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceOrganizationRunTaskGlobalSettingsReadRemovesMissingTask(t *testing.T) {
	testCases := map[string]http.Handler{
		"not found": http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
		}),
		"empty response": http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_, _ = w.Write([]byte(`{}`))
		}),
	}

	for name, handler := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			r := &resourceOrganizationRunTaskGlobalSettings{config: ConfiguredClient{ClientV2: testTfeClientV2(t, handler)}}
			schemaResp := &fwresource.SchemaResponse{}
			r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			diags := state.Set(ctx, &modelDataTFEOrganizationRunTaskGlobalSettings{
				ID:               types.StringValue("task-missing"),
				TaskID:           types.StringValue("task-missing"),
				Enabled:          types.BoolValue(true),
				EnforcementLevel: types.StringValue("mandatory"),
				Stages:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("post_plan")}),
			})
			if diags.HasError() {
				t.Fatalf("failed to build state: %v", diags)
			}

			resp := &fwresource.ReadResponse{State: state}
			r.Read(ctx, fwresource.ReadRequest{State: state}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("read returned diagnostics for an absent task: %v", resp.Diagnostics)
			}
			if !resp.State.Raw.IsNull() {
				t.Fatal("expected missing task settings to be removed from state")
			}
		})
	}
}

func TestGetOrganizationRunTaskConfig(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/example-org/task-configs/for-owner", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q[task-id]"); got != "task-123" {
			t.Errorf("expected task query task-123, got %q", got)
		}
		if got := r.URL.Query().Get("q[owner-id]"); got != "example-org" {
			t.Errorf("expected owner query example-org, got %q", got)
		}
		if got := r.URL.Query().Get("q[owner-type]"); got != "organizations" {
			t.Errorf("expected owner type organizations, got %q", got)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"task-config-123","type":"task-configs","attributes":{"global":true,"allowed-stages":["post_plan"],"enforcement-level":"mandatory"}}}`)
	})

	config, err := getOrganizationRunTaskConfig(ctx, testTfeClientV2(t, mux), "task-123", "example-org")
	if err != nil {
		t.Fatal(err)
	}
	if config == nil || valueOrZero(config.GetId()) != "task-config-123" {
		t.Fatalf("expected task-config-123, got %#v", config)
	}
}

func TestOrganizationRunTaskGlobalTaskConfigUpdateRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/task-configs/task-config-123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var payload struct {
			Data struct {
				Attributes struct {
					Global           *bool    `json:"global"`
					AllowedStages    []string `json:"allowed-stages"`
					EnforcementLevel string   `json:"enforcement-level"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Data.Attributes.Global == nil || *payload.Data.Attributes.Global {
			t.Error("expected global setting to be disabled")
		}
		if got := payload.Data.Attributes.AllowedStages; len(got) != 1 || got[0] != "post_plan" {
			t.Errorf("expected stages to be preserved, got %v", got)
		}
		if got := payload.Data.Attributes.EnforcementLevel; got != "mandatory" {
			t.Errorf("expected enforcement level to be preserved, got %q", got)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprint(w, `{"data":{"id":"task-config-123","type":"task-configs","attributes":{"global":false,"allowed-stages":["post_plan"],"enforcement-level":"mandatory"}}}`)
	})

	enabled := false
	enforcementLevel := "mandatory"
	envelope, err := newOrganizationRunTaskGlobalTaskConfigEnvelope("task-123", "example-org", &enabled, []string{"post_plan"}, &enforcementLevel, false)
	if err != nil {
		t.Fatal(err)
	}
	client := testTfeClientV2(t, mux)
	if _, err := client.API.TaskConfigs().ByTask_config_id("task-config-123").Patch(ctx, envelope, nil); err != nil {
		t.Fatal(err)
	}
}

func TestNewOrganizationRunTaskGlobalTaskConfigEnvelope(t *testing.T) {
	enabled := true
	enforcementLevel := "mandatory"
	envelope, err := newOrganizationRunTaskGlobalTaskConfigEnvelope("task-123", "example-org", &enabled, []string{"pre_plan", "post_apply"}, &enforcementLevel, true)
	if err != nil {
		t.Fatal(err)
	}

	data := envelope.GetData()
	if data == nil || data.GetAttributes() == nil {
		t.Fatal("expected task config data and attributes")
	}
	attributes := data.GetAttributes()
	if !valueOrZero(attributes.GetGlobal()) {
		t.Error("expected global task config to be enabled")
	}
	if got := attributes.GetEnforcementLevel(); got == nil || got.String() != enforcementLevel {
		t.Errorf("expected enforcement level %q, got %v", enforcementLevel, got)
	}
	if got := attributes.GetAllowedStages(); len(got) != 2 || got[0].String() != "pre_plan" || got[1].String() != "post_apply" {
		t.Errorf("unexpected allowed stages: %v", got)
	}

	relationships := data.GetRelationships()
	if relationships == nil || relationships.GetTask() == nil || relationships.GetTask().GetData() == nil {
		t.Fatal("expected task relationship")
	}
	if got := valueOrZero(relationships.GetTask().GetData().GetId()); got != "task-123" {
		t.Errorf("expected task ID task-123, got %q", got)
	}
	if relationships.GetOwner() == nil || relationships.GetOwner().GetData() == nil || relationships.GetOwner().GetData().GetOrganizationsIdentifier() == nil {
		t.Fatal("expected organization owner relationship")
	}
	if got := valueOrZero(relationships.GetOwner().GetData().GetOrganizationsIdentifier().GetId()); got != "example-org" {
		t.Errorf("expected organization example-org, got %q", got)
	}
}

func TestDataModelFromTFEOrganizationRunTaskGlobalTaskConfig(t *testing.T) {
	attributes := models.NewTaskConfigs_attributes()
	enabled := false
	attributes.SetGlobal(&enabled)
	enforcementLevel := models.ADVISORY_TASKCONFIGS_ATTRIBUTES_ENFORCEMENTLEVEL
	attributes.SetEnforcementLevel(&enforcementLevel)
	attributes.SetAllowedStages([]models.TaskConfigs_attributes_allowedStages{
		models.PRE_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
		models.POST_PLAN_TASKCONFIGS_ATTRIBUTES_ALLOWEDSTAGES,
	})
	config := models.NewTaskConfigs()
	config.SetAttributes(attributes)

	result := dataModelFromTFEOrganizationRunTaskGlobalTaskConfig("task-123", config)
	if result.ID.ValueString() != "task-123" || result.TaskID.ValueString() != "task-123" {
		t.Fatalf("expected Terraform identity to remain the task ID, got id=%q task_id=%q", result.ID.ValueString(), result.TaskID.ValueString())
	}
	if result.Enabled.ValueBool() {
		t.Error("expected disabled global task config")
	}
	if result.EnforcementLevel.ValueString() != "advisory" {
		t.Errorf("expected advisory enforcement, got %q", result.EnforcementLevel.ValueString())
	}
	if got := result.Stages.Elements(); len(got) != 2 || got[0].String() != `"pre_plan"` || got[1].String() != `"post_plan"` {
		t.Errorf("unexpected Terraform stages: %v", got)
	}
}

func TestAccTFEOrganizationRunTaskGlobalSettings_validateSchemaAttributeUrl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			// enforcement_level
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters("", `["pre_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute enforcement_level value must be one of: \[.*\]`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters("bad name", `["pre_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute enforcement_level value must be one of: \[.*\]`),
			},
			// stages
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `[]`),
				ExpectError: regexp.MustCompile(`Attribute stages list must contain at least 1 elements.*`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `["pre_plan","BADWOLF","post_plan"]`),
				ExpectError: regexp.MustCompile(`Attribute stages\[1\] value must be.*`),
			},
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_parameters(string(tfe.Mandatory), `["pre_plan","pre_plan","pre_plan"]`),
				ExpectError: regexp.MustCompile(`Error: Duplicate List Value`),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_create(t *testing.T) {
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
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskGlobalEnabled("tfe_organization_run_task.foobar", true),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enforcement_level", "mandatory"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.#", "1"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.0", "post_plan"),
				),
			},
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_update(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskGlobalEnabled("tfe_organization_run_task.foobar", false),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "false"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enforcement_level", "advisory"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.#", "2"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.0", "pre_plan"),
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "stages.1", "post_plan"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_createUnsupported(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createFreeOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
				ExpectError: regexp.MustCompile(`Error: Organization does not support global run tasks`),
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_import(t *testing.T) {
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
		CheckDestroy:             testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrganizationRunTaskGlobalSettings_basic(org.Name, rInt, runTasksURL(), runTasksHMACKey()),
			},
			{
				ResourceName:      "tfe_organization_run_task_global_settings.sut",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/foobar-task-%d", org.Name, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFEOrganizationRunTaskGlobalSettings_Read(t *testing.T) {
	skipUnlessRunTasksDefined(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)
	key := runTasksHMACKey()
	task := createRunTask(t, tfeClient, org.Name, tfe.RunTaskCreateOptions{
		Name:    fmt.Sprintf("tst-task-%s", randomString(t)),
		URL:     runTasksURL(),
		HMACKey: &key,
	})

	org_tf := fmt.Sprintf(`data "tfe_organization" "orgtask" { name = %q }`, org.Name)

	create_settings_tf := fmt.Sprintf(`
		%s
		resource "tfe_organization_run_task_global_settings" "sut" {
			task_id = %q

			enabled           = true
			enforcement_level = "mandatory"
			stages            = ["post_plan"]
		}
		`, org_tf, task.ID)

	delete_task_settings := func() {
		_, err := tfeClient.RunTasks.Update(ctx, task.ID, tfe.RunTaskUpdateOptions{
			Global: &tfe.GlobalRunTaskOptions{
				Enabled: tfe.Bool(false),
			},
		})
		if err != nil {
			t.Fatalf("Error updating task: %s", err)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: create_settings_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
				),
			},
			{
				// Delete the created run task settings and ensure we can re-create it
				PreConfig: delete_task_settings,
				Config:    create_settings_tf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task_global_settings.sut", "enabled", "true"),
				),
			},
			{
				// Delete the created run task settings and ensure we can ignore it if we no longer need to manage it
				PreConfig: delete_task_settings,
				Config:    org_tf,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceNotExist("tfe_organization_run_task_global_settings.sut"),
				),
			},
		},
	})
}

func testAccCheckTFEOrganizationRunTaskGlobalEnabled(resourceName string, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		taskEnvelope, err := testAccConfiguredClient.ClientV2.API.Tasks().ById(rs.Primary.ID).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("error reading Run Task: %w", err)
		}

		if taskEnvelope == nil || taskEnvelope.GetData() == nil {
			return fmt.Errorf("Organization Run Task not found")
		}

		global := taskGlobalConfiguration(taskEnvelope.GetData())
		if global == nil {
			return fmt.Errorf("Organization Run Task exists but does not support global run tasks")
		}

		if enabled := valueOrZero(global.GetEnabled()); enabled != expectedEnabled {
			return fmt.Errorf("Task expected a global enabled value of %t, got %t", expectedEnabled, enabled)
		}

		return nil
	}
}

func testAccTFEOrganizationRunTaskGlobalSettings_basic(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
	organization = "%s"
	url          = "%s"
	name         = "foobar-task-%d"
	enabled      = false
	hmac_key     = "%s"
}

resource "tfe_organization_run_task_global_settings" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  enabled           = true
  enforcement_level = "mandatory"
  stages            = ["post_plan"]
}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}

func testAccTFEOrganizationRunTaskGlobalSettings_parameters(enforceLevel, stages string) string {
	return fmt.Sprintf(`
resource "tfe_organization_run_task" "foobar" {
	organization = "foo"
	url          = "http://somewhere.local"
	name         = "task_name"
	enabled      = false
	hmac_key     = "something"
}

resource "tfe_organization_run_task_global_settings" "sut" {
  task_id = tfe_organization_run_task.foobar.id

  enabled           = true
  enforcement_level = "%s"
  stages            = %s
}
`, enforceLevel, stages)
}

func testAccTFEOrganizationRunTaskGlobalSettings_update(orgName string, rInt int, runTaskURL, runTaskHMACKey string) string {
	return fmt.Sprintf(`
	resource "tfe_organization_run_task" "foobar" {
		organization = "%s"
		url          = "%s"
		name         = "foobar-task-%d-new"
		enabled      = true
		hmac_key     = "%s"
		description  = "a description"
	}

	resource "tfe_organization_run_task_global_settings" "sut" {
		task_id = tfe_organization_run_task.foobar.id

		enabled           = false
		enforcement_level = "advisory"
		stages            = ["pre_plan", "post_plan"]
	}
`, orgName, runTaskURL, rInt, runTaskHMACKey)
}
