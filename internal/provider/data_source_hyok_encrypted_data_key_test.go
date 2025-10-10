package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEHYOKEncryptedDataKeyDataSource_basic(t *testing.T) {
	skipUnlessHYOKEnabled(t)

	hyokEncryptedDataKeyID := os.Getenv("HYOK_ENCRYPTED_DATA_KEY_ID")
	if hyokEncryptedDataKeyID == "" {
		t.Skip("HYOK_ENCRYPTED_DATA_KEY_ID environment variable must be set to run this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEHYOKEncryptedDataKeyDataSourceConfig(hyokEncryptedDataKeyID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_hyok_encrypted_data_key.test", "id", hyokEncryptedDataKeyID),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_encrypted_data_key.test", "encrypted_dek"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_encrypted_data_key.test", "customer_key_name"),
					resource.TestCheckResourceAttrSet("data.tfe_hyok_encrypted_data_key.test", "created_at"),
				),
			},
		},
	})
}

func testAccTFEHYOKEncryptedDataKeyDataSourceConfig(id string) string {
	return `
data "tfe_hyok_encrypted_data_key" "test" {
  id = "` + id + `"
}
`
}
