package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEVaultOIDCConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	orgName := os.Getenv("HYOK_ORGANIZATION_NAME")

	originalAddress := "https://vault.example.com"
	updatedAddress := "https://vault.example2.com"
	originalRoleName := "role-name-1"
	updatedRoleName := "role-name-2"
	originalNamespace := "admin-1"
	updatedNamespace := "admin-2"
	originalAuthPath := "jwt"
	updatedAuthPath := "jwt2"
	originalCACert := ""
	updatedCACert := "some-cert"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVaultOIDCConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVaultOIDCConfigurationConfig(orgName, originalAddress, originalRoleName, originalNamespace, originalAuthPath, originalCACert),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_vault_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "address", originalAddress),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "role_name", originalRoleName),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "namespace", originalNamespace),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "auth_path", originalAuthPath),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "encoded_cacert", originalCACert),
				),
			},
			// Import
			{
				ResourceName:      "tfe_vault_oidc_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccTFEVaultOIDCConfigurationConfig(orgName, updatedAddress, updatedRoleName, updatedNamespace, updatedAuthPath, updatedCACert),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_vault_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "address", updatedAddress),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "role_name", updatedRoleName),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "namespace", updatedNamespace),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "auth_path", updatedAuthPath),
					resource.TestCheckResourceAttr("tfe_vault_oidc_configuration.test", "encoded_cacert", updatedCACert),
				),
			},
		},
	})
}

func testAccTFEVaultOIDCConfigurationConfig(orgName string, address string, roleName string, namespace string, authPath string, cacert string) string {
	return fmt.Sprintf(`
resource "tfe_vault_oidc_configuration" "test" {
	address         =   "%s"
	role_name       =   "%s"
	namespace       =   "%s"
	auth_path       =   "%s"
	encoded_cacert  =   "%s"
	organization    =   "%s"
}
`, address, roleName, namespace, authPath, cacert, orgName)
}

func testAccCheckTFEVaultOIDCConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_vault_oidc_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.VaultOIDCConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("TFE Vault OIDC Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
