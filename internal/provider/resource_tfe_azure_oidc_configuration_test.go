package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEAzureOIDCConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	orgName := os.Getenv("HYOK_ORGANIZATION_NAME")

	originalClientID := "client-id-1"
	updatedClientID := "client-id-2"
	originalSubscriptionID := "subscription-id-1"
	updatedSubscriptionID := "subscription-id-2"
	originalTenantID := "tenant-id-1"
	updatedTenantID := "tenant-id-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAzureOIDCConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAzureOIDCConfigurationConfig(orgName, originalClientID, originalSubscriptionID, originalTenantID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_azure_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "client_id", originalClientID),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "subscription_id", originalSubscriptionID),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "tenant_id", originalTenantID),
				),
			},
			// Import
			{
				ResourceName:      "tfe_azure_oidc_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEAzureOIDCConfigurationConfig(orgName, updatedClientID, updatedSubscriptionID, updatedTenantID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_azure_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "client_id", updatedClientID),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "subscription_id", updatedSubscriptionID),
					resource.TestCheckResourceAttr("tfe_azure_oidc_configuration.test", "tenant_id", updatedTenantID),
				),
			},
		},
	})
}

func testAccTFEAzureOIDCConfigurationConfig(orgName string, clientID string, subscriptionID string, tenantID string) string {
	return fmt.Sprintf(`
resource "tfe_azure_oidc_configuration" "test" {
	client_id           =   "%s"
	subscription_id     =   "%s"
	tenant_id           =   "%s"
	organization        =   "%s"
}
`, clientID, subscriptionID, tenantID, orgName)
}

func testAccCheckTFEAzureOIDCConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_azure_oidc_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.AzureOIDCConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("TFE Azure OIDC Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
