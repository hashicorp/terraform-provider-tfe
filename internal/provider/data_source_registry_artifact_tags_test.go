// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFERegistryArtifactTagsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	module, err := tfeClient.RegistryModules.Create(ctx, org.Name, tfe.RegistryModuleCreateOptions{
		Name:         new(fmt.Sprintf("tst-module-%d", rInt)),
		Provider:     new("test"),
		RegistryName: tfe.PrivateRegistry,
	})
	if err != nil {
		t.Fatalf("error creating registry module: %s", err)
	}

	moduleID := tfe.RegistryModuleID{
		Organization: org.Name,
		Name:         module.Name,
		Provider:     module.Provider,
		RegistryName: module.RegistryName,
		Namespace:    module.Namespace,
	}

	t.Cleanup(func() {
		if err := tfeClient.RegistryModules.DeleteByName(ctx, moduleID); err != nil {
			t.Errorf("error deleting registry module: %s", err)
		}
	})

	if _, err := tfeClient.RegistryModules.Update(ctx, moduleID, tfe.RegistryModuleUpdateOptions{
		TagBindings: []*tfe.TagBinding{
			{Key: "env", Value: "prod"},
			{Key: "team", Value: "platform"},
		},
	}); err != nil {
		t.Fatalf("error adding tag bindings to registry module: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagsDataSourceConfig(
					ArtifactTypeRegistryModule,
					module.ID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "id",
						fmt.Sprintf("%s/%s", ArtifactTypeRegistryModule, module.ID)),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "artifact.type", ArtifactTypeRegistryModule),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "artifact.id", module.ID),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "tags.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.tfe_registry_artifact_tags.foobar", "tags.*", map[string]string{
							"key":   "env",
							"value": "prod",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.tfe_registry_artifact_tags.foobar", "tags.*", map[string]string{
							"key":   "team",
							"value": "platform",
						}),
				),
			},
		},
	})
}

// TestAccTFERegistryArtifactTagsDataSource_noTags verifies that an artifact
// without any tag bindings is not an error and returns an empty tags set.
func TestAccTFERegistryArtifactTagsDataSource_noTags(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createBusinessOrganization(t, tfeClient)
	t.Cleanup(orgCleanup)

	module, err := tfeClient.RegistryModules.Create(ctx, org.Name, tfe.RegistryModuleCreateOptions{
		Name:         new(fmt.Sprintf("tst-module-%d", rInt)),
		Provider:     new("test"),
		RegistryName: tfe.PrivateRegistry,
	})
	if err != nil {
		t.Fatalf("error creating registry module: %s", err)
	}

	t.Cleanup(func() {
		err := tfeClient.RegistryModules.DeleteByName(ctx, tfe.RegistryModuleID{
			Organization: org.Name,
			Name:         module.Name,
			Provider:     module.Provider,
			RegistryName: module.RegistryName,
			Namespace:    module.Namespace,
		})
		if err != nil {
			t.Errorf("error deleting registry module: %s", err)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagsDataSourceConfig(
					ArtifactTypeRegistryModule,
					module.ID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "tags.#", "0"),
					resource.TestCheckResourceAttr(
						"data.tfe_registry_artifact_tags.foobar", "id",
						fmt.Sprintf("%s/%s", ArtifactTypeRegistryModule, module.ID)),
				),
			},
		},
	})
}

func TestAccTFERegistryArtifactTagsDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagsDataSourceConfig(
					ArtifactTypeRegistryModule,
					"mod-notARealArtifact",
				),
				ExpectError: regexp.MustCompile(`Registry Artifact Not Found`),
			},
		},
	})
}

func TestAccTFERegistryArtifactTagsDataSource_invalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFERegistryArtifactTagsDataSourceConfig(
					"unsupported",
					"artifact-id",
				),
				ExpectError: regexp.MustCompile(
					`(?i)(Invalid Attribute Value Match|must be one of)`),
			},
		},
	})
}

func testAccTFERegistryArtifactTagsDataSourceConfig(artifactType string, artifactID string) string {
	return fmt.Sprintf(`
data "tfe_registry_artifact_tags" "foobar" {
  artifact = {
    type = %[1]q
    id   = %[2]q
  }
}
`, artifactType, artifactID)
}
