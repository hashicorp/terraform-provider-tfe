package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEGCPOIDCConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	orgName := os.Getenv("HYOK_ORGANIZATION_NAME")

	originalServiceAccountEmail := "service-account@example.iam.gserviceaccount.com"
	updatedServiceAccountEmail := "updated-service-account@example.iam.gserviceaccount.com"
	originalProjectNumber := "123456789012"
	updatedProjectNumber := "999999999999"
	originalWorkloadProviderName := "workload-provider-name-1"
	updatedWorkloadProviderName := "workload-provider-name-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEGCPOIDCConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEGCPOIDCConfigurationConfig(orgName, originalServiceAccountEmail, originalProjectNumber, originalWorkloadProviderName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_gcp_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "workload_provider_name", originalWorkloadProviderName),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "project_number", originalProjectNumber),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "service_account_email", originalServiceAccountEmail),
				),
			},
			// Import
			{
				ResourceName:      "tfe_gcp_oidc_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEGCPOIDCConfigurationConfig(orgName, updatedServiceAccountEmail, updatedProjectNumber, updatedWorkloadProviderName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_gcp_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "workload_provider_name", updatedWorkloadProviderName),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "project_number", updatedProjectNumber),
					resource.TestCheckResourceAttr("tfe_gcp_oidc_configuration.test", "service_account_email", updatedServiceAccountEmail),
				),
			},
		},
	})
}

func testAccTFEGCPOIDCConfigurationConfig(orgName string, serviceAccountEmail string, projectNumber string, workloadProviderName string) string {
	return fmt.Sprintf(`
resource "tfe_gcp_oidc_configuration" "test" {
	service_account_email   =   "%s"
	project_number          =   "%s"
	workload_provider_name  =   "%s"
	organization            =   "%s"
}
`, serviceAccountEmail, projectNumber, workloadProviderName, orgName)
}

func testAccCheckTFEGCPOIDCConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_gcp_oidc_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.GCPOIDCConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("TFE GCP OIDC Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
