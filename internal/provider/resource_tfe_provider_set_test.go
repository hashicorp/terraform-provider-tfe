// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEProviderSet_basic(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "name", "tst-terraform",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "description", "Provider Set description",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl", "provider \"aws\" {\n\tregion = \"us-east-1\"\n}\n",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "project_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids.#", "1",
					),
				),
			},
			{
				Config: testAccTFEProviderSet_no_global_no_relationship(org.Name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "name", "tst-terraform-updated",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "description", "Provider Set description updated",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/google",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl", "provider \"google\" {\n\tregion = \"us-central1\"\n}\n",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "project_ids",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids",
					),
				),
			},
			{
				Config: testAccTFEProviderSet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "name", "tst-terraform",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "description", "Provider Set description",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl", "provider \"aws\" {\n\tregion = \"us-east-1\"\n}\n",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "project_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids.#", "1",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_update_global_with_relationships(
	t *testing.T,
) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_basic(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "project_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids.#", "1",
					),
				),
			},
			{
				Config: testAccTFEProviderSet_global(org.Name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"tfe_provider_set.foobar",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "true",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "project_ids",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_global(
	t *testing.T,
) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_global(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "true",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "project_ids",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_update_to_global_with_no_relationships(
	t *testing.T,
) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_no_global_no_relationship(org.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "project_ids",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids",
					),
				),
			},
			{
				Config: testAccTFEProviderSet_global(org.Name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"tfe_provider_set.foobar",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "true",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "project_ids",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_minimal(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// when organization is not specified in the config, the provider should use
		// the default organization from the provider configuration, so we need to mux
		// the providers to include that default organization
		ProtoV6ProviderFactories: muxedProvidersWithDefaultOrganization(org.Name),
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_minimal(nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "name", "tst-terraform",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "description", "",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl", "provider \"aws\" {\n\tregion = \"us-east-1\"\n}\n",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "project_ids.#", "0",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids.#", "0",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_update_org_force_recreation(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	orgInitial, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	orgUpdate, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_minimal(&orgInitial.Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
				),
			},
			{
				Config: testAccTFEProviderSet_minimal(&orgUpdate.Name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"tfe_provider_set.foobar",
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_wo(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_wo(org.Name, 1, "us-east-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "name", "tst-terraform",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "description", "",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "global", "false",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl",
					),
					resource.TestCheckNoResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version", "1",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "organization", org.Name,
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "project_ids.#", "0",
					),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "workspace_ids.#", "0",
					),
				),
			},
			{
				// by changing the content of hcl_wo without updating the version, we
				// expect no changes because the provider should ignore changes to
				// hcl_wo when version is set
				Config: testAccTFEProviderSet_wo(org.Name, 1, "us-east-2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccTFEProviderSet_wo(org.Name, 2, "us-east-2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_config_hcl_wo_version", "2",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_validation(t *testing.T) {
	skipUnlessBeta(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEProviderSet_conflict("workspace_ids", "ws-1234123412341234"),
				ExpectError: regexp.MustCompile("workspace_ids cannot be set when global is true"),
			},
			{
				Config:      testAccTFEProviderSet_conflict("project_ids", "prj-1234123412341234"),
				ExpectError: regexp.MustCompile("project_ids cannot be set when global is true"),
			},
			{
				Config:      testAccTFEProviderSet_conflict("workspace_ids", "ws-123"),
				ExpectError: regexp.MustCompile("must be a valid workspace ID"),
			},
			{
				Config:      testAccTFEProviderSet_conflict("project_ids", "prj-123"),
				ExpectError: regexp.MustCompile("must be a valid project ID"),
			},
			{
				Config:      testAccTFEProviderSet_missing_hcl(),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func TestAccTFEProviderSet_provider_source(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	providerSet := &tfe.ProviderSet{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEProviderSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_basic_with_provider_source(org.Name, "hashicorp/aws"),
				ExpectError: regexp.MustCompile(
					"Attribute provider_source must be in the format 'hostname/namespace/type'",
				),
			},
			{
				Config: testAccTFEProviderSet_basic_with_provider_source(org.Name, "registry.terraform.io/hashicorp/aws"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEProviderSetExists("tfe_provider_set.foobar", providerSet),
					resource.TestCheckResourceAttr(
						"tfe_provider_set.foobar", "provider_source", "registry.terraform.io/hashicorp/aws",
					),
				),
			},
		},
	})
}

func TestAccTFEProviderSet_Read(t *testing.T) {
	skipUnlessBeta(t)
	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}
	org, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})
	defer orgCleanup()

	var providerSetID string
	resourceName := "tfe_provider_set.foobar"
	providerSetName := "read-test-" + randomString(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEProviderSet_basic_with_name(org.Name, providerSetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", providerSetName),

					// we capture the provider set ID from the state so that we can use it in
					// the next steps to make out-of-band API calls
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("not found: %s", resourceName)
						}
						providerSetID = rs.Primary.ID
						return nil
					},
				),
			},
			// Step 2: Manually update the name via the API to test synchronization
			{
				PreConfig: func() {
					// We find the ID from the state and update it directly via go-tfe
					_, err := tfeClient.ProviderSets.Update(context.Background(), providerSetID, tfe.ProviderSetUpdateOptions{
						Name: tfe.String(providerSetName + "-updated"),
					})
					if err != nil {
						t.Fatalf("error updating provider set out-of-band: %v", err)
					}
				},
				Config: testAccTFEProviderSet_basic_with_name(org.Name, providerSetName),
				// Terraform should run Read, see the "-updated" name, and then plan to change it back to providerSetName
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// Step 3: Manually delete via the API to test out-of-band deletion
			{
				PreConfig: func() {
					err := tfeClient.ProviderSets.Delete(context.Background(), providerSetID)
					if err != nil {
						t.Fatalf("error deleting provider set out-of-band: %v", err)
					}
				},
				Config:             testAccTFEProviderSet_basic_with_name(org.Name, providerSetName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true, // Terraform should see it's gone and plan a "Create"
			},
		},
	})
}

func testAccCheckTFEProviderSetExists(n string, providerSet *tfe.ProviderSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ps, err := testAccConfiguredClient.Client.ProviderSets.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if ps.ID != rs.Primary.ID {
			return fmt.Errorf("ProviderSet not found")
		}

		*providerSet = *ps

		return nil
	}
}

func testAccCheckTFEProviderSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_provider_set" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.ProviderSets.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Provider set %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func TestTFEProviderSetNotImportableInPrepStub(t *testing.T) {
	if _, ok := any(&resourceTFEProviderSet{}).(fwresource.ResourceWithImportState); ok {
		t.Fatal("expected provider set prep stub to not support import")
	}
}

func TestFrameworkProviderResources_includeProviderSet(t *testing.T) {
	resources := (&frameworkProvider{}).Resources(context.Background())

	found := false
	for _, constructor := range resources {
		if _, ok := constructor().(*resourceTFEProviderSet); ok {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected framework provider resources to include tfe_provider_set")
	}
}

func testAccTFEProviderSet_basic(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_project" "foo" {
  name         = "project-foo"
  organization = local.organization_name
}

resource "tfe_provider_set" "foobar" {
  name                = "tst-terraform"
  description         = "Provider Set description"
  organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/aws"
	global              = false
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT

  project_ids =   [ tfe_project.foo.id ]
  workspace_ids = [ tfe_workspace.foo.id ]
}`, organization)
}

func testAccTFEProviderSet_global(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_provider_set" "foobar" {
  name                = "tst-terraform"
  description         = "Provider Set description"
  organization        = local.organization_name
	provider_source     = "registry.terraform.io/hashicorp/aws"
	global              = true
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}`, organization)
}

func testAccTFEProviderSet_no_global_no_relationship(organization string) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo"
  organization = local.organization_name
}

resource "tfe_project" "foo" {
  name         = "project-foo"
  organization = local.organization_name
}

resource "tfe_provider_set" "foobar" {
  name                = "tst-terraform-updated"
  description         = "Provider Set description updated"
  organization        = local.organization_name
  provider_source     = "registry.terraform.io/hashicorp/google"
  global              = false
  provider_config_hcl = <<-EOT
provider "google" {
	region = "us-central1"
}
EOT
}`, organization)
}

func testAccTFEProviderSet_minimal(organization *string) string {
	orgConfig := "// should rely on provider for organization"
	if organization != nil {
		orgConfig = fmt.Sprintf("organization        = %q", *organization)
	}
	return fmt.Sprintf(`
resource "tfe_provider_set" "foobar" {
	%s
  name                = "tst-terraform"
	provider_source     = "registry.terraform.io/hashicorp/aws"
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}`, orgConfig)
}

func testAccTFEProviderSet_missing_hcl() string {
	return `
resource "tfe_provider_set" "foobar" {
	organization        = "my-org"
  name                = "tst-terraform"
	provider_source     = "registry.terraform.io/hashicorp/aws"
}`
}

func testAccTFEProviderSet_wo(organization string, version int64, region string) string {
	return fmt.Sprintf(`
locals {
	organization_name = "%s"
	version           = %d
	region            = "%s"
}

resource "tfe_provider_set" "foobar" {
  name                           = "tst-terraform"
	organization                   = local.organization_name
	provider_source                = "registry.terraform.io/hashicorp/aws"
	provider_config_hcl_wo_version = local.version
	provider_config_hcl_wo         = <<-EOT
provider "aws" {
	region = local.region
}
EOT
}`, organization, version, region)
}

func testAccTFEProviderSet_conflict(relationship, id string) string {
	return fmt.Sprintf(`
resource "tfe_provider_set" "error" {
  name            = "conflict-test"
  organization    = "my-org"
  provider_source = "registry.terraform.io/hashicorp/aws"
  global          = true
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
  %s   = ["%s"] # This should trigger the validation error
}`, relationship, id)
}

func testAccTFEProviderSet_basic_with_name(organization, name string) string {
	return fmt.Sprintf(`
locals {
	organization_name = "%s"
	name              = "%s"
}
resource "tfe_provider_set" "foobar" {
  name                           = local.name
	organization                   = local.organization_name
	provider_source                = "registry.terraform.io/hashicorp/aws"
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}`, organization, name)
}

func testAccTFEProviderSet_basic_with_provider_source(organization, providerSource string) string {
	return fmt.Sprintf(`
locals {
	organization_name = "%s"
	provider_source   = "%s"
}
resource "tfe_provider_set" "foobar" {
  name                           = "provider-source-test"
	organization                   = local.organization_name
	provider_source                = local.provider_source
	provider_config_hcl = <<-EOT
provider "aws" {
	region = "us-east-1"
}
EOT
}`, organization, providerSource)
}
