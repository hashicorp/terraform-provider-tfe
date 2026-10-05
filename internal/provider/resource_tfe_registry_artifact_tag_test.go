// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFERegistryArtifactTag_module(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERegistryArtifactTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagModuleConfig(
					org.Name,
					rInt,
					"first",
					testAccRegistryArtifactTagsEnv,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryModule,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.first",
						ArtifactTypeRegistryModule,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"env": "prod",
						},
					),
				),
			},
			{
				Config: testAccTFERegistryArtifactTagModuleConfig(
					org.Name,
					rInt,
					"first",
					testAccRegistryArtifactTagsTeam,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryModule,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.first",
						ArtifactTypeRegistryModule,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"team": "platform",
						},
					),
				),
			},
			// Change the target artifact only, holding tags constant, so this step
			// isolates replacement.
			{
				Config: testAccTFERegistryArtifactTagModuleConfig(
					org.Name,
					rInt,
					"second",
					testAccRegistryArtifactTagsTeam,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"tfe_registry_artifact_tag.test",
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryModule,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.first",
						ArtifactTypeRegistryModule,
						nil,
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.second",
						ArtifactTypeRegistryModule,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"team": "platform",
						},
					),
				),
			},
			// Change the tags only, holding the target artifact constant and
			// assert no replacement occurs.
			{
				Config: testAccTFERegistryArtifactTagModuleConfig(
					org.Name,
					rInt,
					"second",
					testAccRegistryArtifactTagsEnvAndTeam,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"tfe_registry_artifact_tag.test",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryModule,
						map[string]string{
							"env":  "prod",
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.second",
						ArtifactTypeRegistryModule,
						map[string]string{
							"env":  "prod",
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"env":  "prod",
							"team": "platform",
						},
					),
				),
			},
			{
				ResourceName: "tfe_registry_artifact_tag.test",
				ImportState:  true,
				ImportStateIdFunc: testAccRegistryArtifactTagImportStateID(
					"tfe_registry_artifact_tag.test",
				),
				ImportStateVerify: true,
			},
			{
				Config: testAccTFERegistryArtifactTagModuleConfig(
					org.Name,
					rInt,
					"second",
					"[]",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryModule,
						nil,
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_module.second",
						ArtifactTypeRegistryModule,
						nil,
					),
					testAccCheckTFERegistryArtifactTagsDataSource(nil),
				),
			},
			{
				ResourceName: "tfe_registry_artifact_tag.test",
				ImportState:  true,
				ImportStateIdFunc: testAccRegistryArtifactTagImportStateID(
					"tfe_registry_artifact_tag.test",
				),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFERegistryArtifactTag_provider(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERegistryArtifactTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagProviderConfig(
					org.Name,
					rInt,
					testAccRegistryArtifactTagsEnv,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryProvider,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_provider.test",
						ArtifactTypeRegistryProvider,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"env": "prod",
						},
					),
				),
			},
			{
				Config: testAccTFERegistryArtifactTagProviderConfig(
					org.Name,
					rInt,
					testAccRegistryArtifactTagsTeam,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryProvider,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagBindingsForResource(
						"tfe_registry_provider.test",
						ArtifactTypeRegistryProvider,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"team": "platform",
						},
					),
				),
			},
			{
				ResourceName: "tfe_registry_artifact_tag.test",
				ImportState:  true,
				ImportStateIdFunc: testAccRegistryArtifactTagImportStateID(
					"tfe_registry_artifact_tag.test",
				),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFERegistryArtifactTag_component(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	component, err := tfeClient.RegistryComponents.Create(
		ctx,
		org.Name,
		tfe.RegistryComponentCreateOptions{
			Name: fmt.Sprintf("tst-component-%d", rInt),
		},
	)
	if err != nil {
		t.Fatalf("error creating registry component: %s", err)
	}

	t.Cleanup(func() {
		err := tfeClient.RegistryComponents.Delete(ctx, component.ID)
		if err != nil && !errors.Is(err, tfe.ErrResourceNotFound) {
			t.Errorf("error deleting registry component: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFERegistryArtifactTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagComponentConfig(
					org.Name,
					rInt,
					component.ID,
					testAccRegistryArtifactTagsEnv,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryComponent,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagBindings(
						component.ID,
						ArtifactTypeRegistryComponent,
						map[string]string{
							"env": "prod",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"env": "prod",
						},
					),
				),
			},
			{
				Config: testAccTFERegistryArtifactTagComponentConfig(
					org.Name,
					rInt,
					component.ID,
					testAccRegistryArtifactTagsTeam,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTFERegistryArtifactTagState(
						"tfe_registry_artifact_tag.test",
						ArtifactTypeRegistryComponent,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagBindings(
						component.ID,
						ArtifactTypeRegistryComponent,
						map[string]string{
							"team": "platform",
						},
					),
					testAccCheckTFERegistryArtifactTagsDataSource(
						map[string]string{
							"team": "platform",
						},
					),
				),
			},
			{
				ResourceName: "tfe_registry_artifact_tag.test",
				ImportState:  true,
				ImportStateIdFunc: testAccRegistryArtifactTagImportStateID(
					"tfe_registry_artifact_tag.test",
				),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTFERegistryArtifactTag_validation(t *testing.T) {
	testCases := map[string]struct {
		config      string
		expectError *regexp.Regexp
	}{
		"invalid artifact type": {
			config: testAccTFERegistryArtifactTagValidationConfig(
				"unsupported",
				testAccRegistryArtifactTagsEnv,
			),
			expectError: regexp.MustCompile(
				`(?i)(Invalid Attribute Value Match|must be one of)`,
			),
		},
		"empty tag key": {
			config: testAccTFERegistryArtifactTagValidationConfig(
				ArtifactTypeRegistryModule,
				`[{"key" = "", "value" = "prod"}]`,
			),
			expectError: regexp.MustCompile(
				`(?i)(Invalid Attribute Value Length|at least 1)`,
			),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(t)
				},
				ProtoV6ProviderFactories: testAccMuxedProviders,
				Steps: []resource.TestStep{
					{
						Config:      testCase.config,
						ExpectError: testCase.expectError,
					},
				},
			})
		})
	}
}

func TestAccTFERegistryArtifactTag_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagValidationConfig(
					ArtifactTypeRegistryModule,
					testAccRegistryArtifactTagsEnv,
				),
				ExpectError: regexp.MustCompile(
					`Registry Artifact Not Found`,
				),
			},
		},
	})
}

func TestResourceTFERegistryArtifactTagImportState_rejectsInvalidID(
	t *testing.T,
) {
	testCases := map[string]struct {
		importID string
		summary  string
	}{
		"missing separator": {
			importID: "registry-module",
			summary:  "Invalid Import ID Format",
		},
		"unsupported type": {
			importID: "unsupported/artifact-id",
			summary:  "Invalid Artifact Type",
		},
		"empty artifact type": {
			importID: "/artifact-id",
			summary:  "Invalid Import ID Format",
		},
		"empty artifact identity": {
			importID: "registry-module/",
			summary:  "Invalid Import ID Format",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := &frameworkresource.ImportStateResponse{}

			(&resourceTFERegistryArtifactTag{}).ImportState(
				ctx,
				frameworkresource.ImportStateRequest{
					ID: testCase.importID,
				},
				resp,
			)

			if !resp.Diagnostics.HasError() {
				t.Fatalf(
					"expected import ID %q to produce an error",
					testCase.importID,
				)
			}

			summary := resp.Diagnostics.Errors()[0].Summary()
			if summary != testCase.summary {
				t.Fatalf(
					"expected diagnostic summary %q, got %q",
					testCase.summary,
					summary,
				)
			}
		})
	}
}

func TestRegistryArtifactTagBindingModelConversions(t *testing.T) {
	bindings := modelTagsToTFETagBindings([]modelTag{
		{
			Key:   types.StringValue("env"),
			Value: types.StringValue("prod"),
		},
		{
			Key:   types.StringNull(),
			Value: types.StringValue("ignored"),
		},
		{
			Key:   types.StringValue(""),
			Value: types.StringValue("ignored"),
		},
		{
			Key:   types.StringValue("team"),
			Value: types.StringValue("platform"),
		},
	})

	expected := map[string]string{
		"env":  "prod",
		"team": "platform",
	}

	if diff := diffRegistryArtifactTagBindings(bindings, expected); diff != "" {
		t.Fatal(diff)
	}

	model := modelFromTFETagBindings(append(bindings, nil))
	if len(model) != 2 {
		t.Fatalf("expected two model tags, got %d", len(model))
	}

	if model[0].Key.ValueString() != "env" ||
		model[0].Value.ValueString() != "prod" {
		t.Fatalf("unexpected first model tag: %#v", model[0])
	}

	if model[1].Key.ValueString() != "team" ||
		model[1].Value.ValueString() != "platform" {
		t.Fatalf("unexpected second model tag: %#v", model[1])
	}
}

func testAccCheckTFERegistryArtifactTagState(resourceName string, artifactType string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		artifactID := rs.Primary.Attributes["artifact.id"]
		if artifactID == "" {
			return fmt.Errorf("%s has no artifact.id", resourceName)
		}

		expectedID := fmt.Sprintf("%s/%s", artifactType, artifactID)
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("%s has ID %q, expected %q", resourceName, rs.Primary.ID, expectedID)
		}

		actualType := rs.Primary.Attributes["artifact.type"]
		if actualType != artifactType {
			return fmt.Errorf("%s has artifact.type %q, expected %q", resourceName, actualType, artifactType)
		}

		return testAccCheckRegistryArtifactTagsInState(rs, expected)
	}
}

func testAccCheckTFERegistryArtifactTagsDataSource(expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.tfe_registry_artifact_tags.test"]
		if !ok {
			return fmt.Errorf("data source not found: data.tfe_registry_artifact_tags.test")
		}

		managed, ok := s.RootModule().Resources["tfe_registry_artifact_tag.test"]
		if !ok {
			return fmt.Errorf("resource not found: tfe_registry_artifact_tag.test")
		}

		dataSourceType := rs.Primary.Attributes["artifact.type"]
		dataSourceID := rs.Primary.Attributes["artifact.id"]
		managedType := managed.Primary.Attributes["artifact.type"]
		managedID := managed.Primary.Attributes["artifact.id"]

		if dataSourceType != managedType || dataSourceID != managedID {
			return fmt.Errorf("data source artifact %q/%q does not match managed artifact %q/%q", dataSourceType, dataSourceID, managedType, managedID)
		}

		expectedID := fmt.Sprintf("%s/%s", managedType, managedID)
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("data source has ID %q, expected %q", rs.Primary.ID, expectedID)
		}

		return testAccCheckRegistryArtifactTagsInState(rs, expected)
	}
}

func testAccCheckRegistryArtifactTagsInState(rs *terraform.ResourceState, expected map[string]string) error {
	actual := make(map[string]string)
	for attribute, key := range rs.Primary.Attributes {
		if !strings.HasPrefix(attribute, "tags.") ||
			!strings.HasSuffix(attribute, ".key") {
			continue
		}
		// eg: prefix will have the value of "tags.0." if the attribute is "tags.0.key"
		prefix := strings.TrimSuffix(attribute, "key")
		actual[key] = rs.Primary.Attributes[prefix+"value"]
	}

	if diff := diffRegistryArtifactTagMap(actual, expected); diff != "" {
		return fmt.Errorf("%s state: %s", rs.Type, diff)
	}

	return nil
}

func testAccCheckTFERegistryArtifactTagBindingsForResource(resourceName string, artifactType string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		return checkRegistryArtifactTagBindings(rs.Primary.ID, artifactType, expected)
	}
}

func testAccCheckTFERegistryArtifactTagBindings(artifactID string, artifactType string, expected map[string]string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return checkRegistryArtifactTagBindings(
			artifactID,
			artifactType,
			expected,
		)
	}
}

func checkRegistryArtifactTagBindings(artifactID string, artifactType string, expected map[string]string) error {
	bindings, err := testAccListRegistryArtifactTagBindings(artifactType, artifactID)
	if err != nil {
		return fmt.Errorf("error reading tag bindings for %s %q: %w", artifactType, artifactID, err)
	}

	if diff := diffRegistryArtifactTagBindings(bindings, expected); diff != "" {
		return fmt.Errorf("%s %q: %s", artifactType, artifactID, diff)
	}

	return nil
}

func testAccListRegistryArtifactTagBindings(artifactType string, artifactID string) ([]*tfe.TagBinding, error) {
	switch artifactType {
	case ArtifactTypeRegistryModule:
		return testAccConfiguredClient.Client.RegistryModules.
			ListTagBindings(ctx, artifactID)

	case ArtifactTypeRegistryProvider:
		return testAccConfiguredClient.Client.RegistryProviders.
			ListTagBindings(ctx, artifactID)

	case ArtifactTypeRegistryComponent:
		return testAccConfiguredClient.Client.RegistryComponents.
			ListTagBindings(ctx, artifactID)

	default:
		return nil, fmt.Errorf(
			"unsupported registry artifact type %q",
			artifactType,
		)
	}
}

func diffRegistryArtifactTagBindings(actual []*tfe.TagBinding, expected map[string]string) string {
	actualMap := make(map[string]string, len(actual))

	for _, binding := range actual {
		if binding != nil {
			actualMap[binding.Key] = binding.Value
		}
	}

	return diffRegistryArtifactTagMap(actualMap, expected)
}

func diffRegistryArtifactTagMap(actual map[string]string, expected map[string]string) string {
	if len(actual) != len(expected) {
		return fmt.Sprintf("got tags %#v, expected %#v", actual, expected)
	}

	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok || actualValue != expectedValue {
			return fmt.Sprintf("got tags %#v, expected %#v", actual, expected)
		}
	}

	return ""
}

func testAccCheckTFERegistryArtifactTagDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_registry_artifact_tag" {
			continue
		}

		artifactType := rs.Primary.Attributes["artifact.type"]
		artifactID := rs.Primary.Attributes["artifact.id"]

		bindings, err := testAccListRegistryArtifactTagBindings(
			artifactType,
			artifactID,
		)
		if errors.Is(err, tfe.ErrResourceNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf(
				"error reading tag bindings for %s %q: %w",
				artifactType,
				artifactID,
				err,
			)
		}

		if len(bindings) > 0 {
			return fmt.Errorf(
				"tag bindings still exist for %s %q: %#v",
				artifactType,
				artifactID,
				bindings,
			)
		}
	}

	return nil
}

func testAccRegistryArtifactTagImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		artifactType := rs.Primary.Attributes["artifact.type"]
		artifactID := rs.Primary.Attributes["artifact.id"]

		if artifactType == "" || artifactID == "" {
			return "", fmt.Errorf("%s has no artifact.type or artifact.id", resourceName)
		}

		return fmt.Sprintf("%s/%s", artifactType, artifactID), nil
	}
}

const (
	testAccRegistryArtifactTagsEnv = `[
  {
    key   = "env"
    value = "prod"
  }
]`

	testAccRegistryArtifactTagsTeam = `[
  {
    key   = "team"
    value = "platform"
  }
]`

	testAccRegistryArtifactTagsEnvAndTeam = `[
  {
    key   = "env"
    value = "prod"
  },
  {
    key   = "team"
    value = "platform"
  }
]`
)

func testAccTFERegistryArtifactTagModuleConfig(orgName string, rInt int, target string, tags string) string {
	return fmt.Sprintf(`
resource "tfe_project" "tag_catalog" {
  organization = %[1]q
  name         = "tst-tag-catalog-%[2]d"

  tags = {
    env  = "prod"
    team = "platform"
  }
}

resource "tfe_registry_module" "first" {
  organization    = %[1]q
  name            = "tst-module-first-%[2]d"
  module_provider = "test"
  registry_name   = "private"
}

resource "tfe_registry_module" "second" {
  organization    = %[1]q
  name            = "tst-module-second-%[2]d"
  module_provider = "test"
  registry_name   = "private"
}

resource "tfe_registry_artifact_tag" "test" {
  artifact = {
    type = "registry-module"
    id   = tfe_registry_module.%[3]s.id
  }

  tags = %[4]s

  depends_on = [tfe_project.tag_catalog]
}

data "tfe_registry_artifact_tags" "test" {
  artifact = tfe_registry_artifact_tag.test.artifact

  depends_on = [tfe_registry_artifact_tag.test]
}
`, orgName, rInt, target, tags)
}

func testAccTFERegistryArtifactTagProviderConfig(orgName string, rInt int, tags string) string {
	return fmt.Sprintf(`
resource "tfe_project" "tag_catalog" {
  organization = %[1]q
  name         = "tst-tag-catalog-%[2]d"

  tags = {
    env  = "prod"
    team = "platform"
  }
}

resource "tfe_registry_provider" "test" {
  organization = %[1]q
  name         = "tst-provider-%[2]d"
}

resource "tfe_registry_artifact_tag" "test" {
  artifact = {
    type = "registry-provider"
    id   = tfe_registry_provider.test.id
  }

  tags = %[3]s

  depends_on = [tfe_project.tag_catalog]
}

data "tfe_registry_artifact_tags" "test" {
  artifact = tfe_registry_artifact_tag.test.artifact

  depends_on = [tfe_registry_artifact_tag.test]
}
`, orgName, rInt, tags)
}

func testAccTFERegistryArtifactTagComponentConfig(orgName string, rInt int, componentID string, tags string) string {
	return fmt.Sprintf(`
resource "tfe_project" "tag_catalog" {
  organization = %[1]q
  name         = "tst-tag-catalog-%[2]d"

  tags = {
    env  = "prod"
    team = "platform"
  }
}

resource "tfe_registry_artifact_tag" "test" {
  artifact = {
    type = "registry-component"
    id   = %[3]q
  }

  tags = %[4]s

  depends_on = [tfe_project.tag_catalog]
}

data "tfe_registry_artifact_tags" "test" {
  artifact = tfe_registry_artifact_tag.test.artifact

  depends_on = [tfe_registry_artifact_tag.test]
}
`, orgName, rInt, componentID, tags)
}

func testAccTFERegistryArtifactTagValidationConfig(artifactType string, tags string) string {
	return fmt.Sprintf(`
resource "tfe_registry_artifact_tag" "test" {
  artifact = {
    type = %q
    id   = "artifact-id"
  }

  tags = %s
}
`, artifactType, tags)
}
