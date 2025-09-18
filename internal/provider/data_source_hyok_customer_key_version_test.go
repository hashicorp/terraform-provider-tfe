package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEHYOKCustomerKeyVersionDataSource_basic(t *testing.T) {
	hyokCustomerKeyVersionId := os.Getenv("HYOK_CUSTOMER_KEY_VERSION_ID")
	if hyokCustomerKeyVersionId == "" {
		t.Skip("HYOK_CUSTOMER_KEY_VERSION_ID environment variable must be set to run this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTFEHYOKCustomerKeyVersionDataSourceConfig(hyokCustomerKeyVersionId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_hyok_customer_key_version.test", "id", hyokCustomerKeyVersionId),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "status"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "key_version"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "workspaces_secured"),
				),
			},
		},
	})
}

func testAccTFEHYOKCustomerKeyVersionDataSourceConfig(id string) string {
	return `
data "tfe_hyok_customer_key_version" "test" {
  id = "` + id + `"
}
`
}
