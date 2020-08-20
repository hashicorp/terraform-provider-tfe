package tfe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccTFECurrentRunDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if TFE_RUN_ID == "" {
				t.Skip("Please set TFE_RUN_ID to run this test")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFECurrentRunDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_current_run.foobar", "id", TFE_RUN_ID),
					resource.TestCheckResourceAttrSet(
						"data.tfe_current_run.foobar", "workspace.0.id"),
				),
			},
		},
	})
}

func testAccTFECurrentRunDataSourceConfig() string {
	return fmt.Sprintf(`
data "tfe_current_run" "foobar" {
}`)
}
