// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFERegistryArtifactTag_basicModule(t *testing.T) {
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
				Config: testAccTFERegistryArtifactTag_basicModule(org.Name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFERegistryArtifactTagExists("tfe_registry_artifact_tag.test"),
					resource.TestCheckResourceAttrSet("tfe_registry_artifact_tag.test", "id"),
					resource.TestCheckResourceAttr("tfe_registry_artifact_tag.test", "artifact.type", ArtifactTypeRegistryModule),
					resource.TestCheckResourceAttr("tfe_registry_artifact_tag.test", "tags.0.key", "env"),
					resource.TestCheckResourceAttr("tfe_registry_artifact_tag.test", "tags.0.value", "prod"),
				),
			},
			{
				ResourceName:      "tfe_registry_artifact_tag.test",
				ImportState:       true,
				ImportStateIdFunc: testAccRegistryArtifactTagImportStateID("tfe_registry_artifact_tag.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFERegistryArtifactTagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		artifactID := rs.Primary.Attributes["artifact.id"]
		if artifactID == "" {
			return fmt.Errorf("artifact.id is not set")
		}

		bindings, err := testAccConfiguredClient.Client.RegistryModules.ListTagBindings(ctx, artifactID)
		if err != nil {
			return fmt.Errorf("error reading tag bindings for artifact %s: %w", artifactID, err)
		}

		if len(bindings) == 0 {
			return fmt.Errorf("no tag bindings found for artifact %s", artifactID)
		}

		return nil
	}
}

func testAccCheckTFERegistryArtifactTagDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_registry_artifact_tag" {
			continue
		}

		artifactID := rs.Primary.Attributes["artifact.id"]

		bindings, err := testAccConfiguredClient.Client.RegistryModules.ListTagBindings(ctx, artifactID)
		if err != nil {
			if errors.Is(err, tfe.ErrResourceNotFound) {
				continue
			}
			return fmt.Errorf("error reading tag bindings for artifact %s: %w", artifactID, err)
		}

		if len(bindings) > 0 {
			return fmt.Errorf("tag bindings still exist for artifact %s", artifactID)
		}
	}
	return nil
}

func testAccRegistryArtifactTagImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("not found: %s", n)
		}

		artifactType := rs.Primary.Attributes["artifact.type"]
		artifactID := rs.Primary.Attributes["artifact.id"]

		if artifactType == "" || artifactID == "" {
			return "", fmt.Errorf("artifact.type or artifact.id is not set")
		}

		return fmt.Sprintf("%s/%s", artifactType, artifactID), nil
	}
}

func testAccTFERegistryArtifactTag_basicModule(orgName string, rInt int) string {
	return fmt.Sprintf(`
resource "tfe_registry_module" "test" {
	organization  = "%s"
	name          = "tst-module-%d"
	module_provider = "my_provider"
	registry_name = "private"
}

resource "tfe_registry_artifact_tag" "test" {
	artifact = {
		type = "registry-module"
		id   = tfe_registry_module.test.id
	}
	tags = [
		{
			key   = "env"
			value = "prod"
		}
	]
}
`, orgName, rInt)
}
