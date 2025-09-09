package provider

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEHYOKCustomerKeyVersionDataSource_basic(t *testing.T) {
	//tfeClient, err := getClientUsingEnv()
	//if err != nil {
	//	t.Fatal(err)
	//}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTFEHYOKCustomerKeyVersionDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_hyok_customer_key_version.test", "id", "keyv-BWZTzt2J75DsdwH8"),
					resource.TestCheckResourceAttr("data.tfe_hyok_customer_key_version.test", "status", "available"),
					resource.TestCheckResourceAttr("data.tfe_hyok_customer_key_version.test", "key_version", "1"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_customer_key_version.test", "updated_at"),
				),
			},
		},
	})
}

const testAccTFEHYOKCustomerKeyVersionDataSourceConfig = `
data "tfe_organization" "org" {
  name = "dretli-hyok-org"
}

data "tfe_hyok_customer_key_version" "test" {
  id = "keyv-BWZTzt2J75DsdwH8"
}
`
