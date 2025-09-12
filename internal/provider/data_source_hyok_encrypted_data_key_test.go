package provider

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEHYOKEncryptedDataKeyDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTFEHYOKEncryptedDataKeyDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_hyok_encrypted_data_key.test", "id", "dek-wuLiejfGtNLLuiH9"),
					resource.TestCheckResourceAttr("data.tfe_hyok_encrypted_data_key.test", "encrypted_dek", "dmF1bHQ6djEwOjdFb3gzNERXQ05zNGVNelNSb09waWp3dGE4SmlNa0JjWFRsQ25KbXlRNlZWRGpCbnFtOFBvbGkvb1ZGTkQ3UVFybDNoNzBrT2hScnlHUlZS"),
					resource.TestCheckResourceAttr("data.tfe_hyok_encrypted_data_key.test", "customer_key_name", "tf-rocket-hyok-oasis"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_encrypted_data_key.test", "created_at"),
				),
			},
		},
	})
}

const testAccTFEHYOKEncryptedDataKeyDataSourceConfig = `
data "tfe_organization" "org" {
  name = "dretli-hyok-org"
}

data "tfe_hyok_encrypted_data_key" "test" {
  id = "dek-wuLiejfGtNLLuiH9"
}
`
