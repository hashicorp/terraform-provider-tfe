package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTFEAWSOIDCConfiguration_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	orgName := os.Getenv("HYOK_ORGANIZATION_NAME")

	originalRoleARN := "arn:aws:iam::123456789012:role/terraform-provider-tfe-example-1"
	newRoleARN := "arn:aws:iam::123456789012:role/terraform-provider-tfe-example-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEAWSOIDCConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEAWSOIDCConfigurationConfig(orgName, originalRoleARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_aws_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_aws_oidc_configuration.test", "role_arn", originalRoleARN),
				),
			},
			// Import
			{
				ResourceName:      "tfe_aws_oidc_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update role ARN
			{
				Config: testAccTFEAWSOIDCConfigurationConfig(orgName, newRoleARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tfe_aws_oidc_configuration.test", "id"),
					resource.TestCheckResourceAttr("tfe_aws_oidc_configuration.test", "role_arn", newRoleARN),
				),
			},
		},
	})
}

func testAccTFEAWSOIDCConfigurationConfig(orgName string, roleARN string) string {
	return fmt.Sprintf(`
resource "tfe_aws_oidc_configuration" "test" {
	role_arn    = "%s"
	organization = "%s"
}
`, roleARN, orgName)
}

func testAccCheckTFEAWSOIDCConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_aws_oidc_configuration" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.AWSOIDCConfigurations.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("TFE AWS OIDC Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
