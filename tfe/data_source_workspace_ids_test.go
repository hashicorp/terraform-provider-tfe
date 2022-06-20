package tfe

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEWorkspaceIDsDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.0", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.1", fmt.Sprintf("workspace-bar-%d", rInt)),

					// organization attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "organization", orgName),

					// full_names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "full_names.%", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("full_names.workspace-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("full_names.workspace-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-bar-%d", rInt, rInt),
					),

					// ids attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),

					// id attribute
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_wildcard(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "names.0", "*"),

					// organization attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "organization", orgName),

					// full_names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "full_names.%", "3"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("full_names.workspace-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("full_names.workspace-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-bar-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar",
						fmt.Sprintf("full_names.workspace-dummy-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-dummy-%d", rInt, rInt),
					),

					// ids attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.foobar", "ids.%", "3"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("ids.workspace-bar-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.foobar", fmt.Sprintf("ids.workspace-dummy-%d", rInt)),

					// id attribute
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.foobar", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_tags(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_tags(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "organization", orgName),

					// full_names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "full_names.%", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good",
						fmt.Sprintf("full_names.workspace-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good",
						fmt.Sprintf("full_names.workspace-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-bar-%d", rInt, rInt),
					),

					// ids attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "ids.%", "2"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", fmt.Sprintf("ids.workspace-bar-%d", rInt)),

					// id attribute
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.good", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_searchByTagAndName(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_searchByTagAndName(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					// organization attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "organization", orgName),

					// full_names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "full_names.%", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good",
						fmt.Sprintf("full_names.workspace-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-foo-%d", rInt, rInt),
					),

					// ids attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "ids.%", "1"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", fmt.Sprintf("ids.workspace-foo-%d", rInt)),

					// id attribute
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.good", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_empty(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEWorkspaceIDsDataSourceConfig_empty(rInt),
				ExpectError: regexp.MustCompile("one of `names,tag_names` must be specified"),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_namesEmpty(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_namesEmpty(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "names.#", "2"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "names.0", ""),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "names.1", fmt.Sprintf("workspace-foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "organization", orgName),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", "full_names.%"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "full_names.%", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good",
						fmt.Sprintf("full_names.workspace-foo-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-foo-%d", rInt, rInt),
					),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "ids.%", "1"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.good", "ids.%"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.good", "id"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", fmt.Sprintf("ids.workspace-foo-%d", rInt)),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_excludeTags(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_excludeTags(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "organization", orgName),

					// full_names attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "full_names.%", "1"),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good",
						fmt.Sprintf("full_names.workspace-bar-%d", rInt),
						fmt.Sprintf("tst-terraform-%d/workspace-bar-%d", rInt, rInt),
					),

					// ids attribute
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "ids.%", "1"),
					resource.TestCheckResourceAttrSet(
						"data.tfe_workspace_ids.good", fmt.Sprintf("ids.workspace-bar-%d", rInt)),

					// id attribute
					resource.TestCheckResourceAttrSet("data.tfe_workspace_ids.good", "id"),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceIDsDataSource_sameTagInTagNamesAndExcludeTags(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceIDsDataSourceConfig_sameTagInTagNamesAndExcludeTags(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "organization", orgName),

					// full_names attribute should be empty
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "full_names.%", "0"),

					// ids attribute should be empty
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_ids.good", "ids.%", "0"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceIDsDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_workspace_ids" "foobar" {
  names        = [tfe_workspace.foo.name, tfe_workspace.bar.name]
  organization = tfe_organization.foobar.name
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_wildcard(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_workspace_ids" "foobar" {
  names        = ["*"]
  organization = tfe_workspace.dummy.organization
  depends_on = [
    tfe_workspace.foo,
    tfe_workspace.bar,
    tfe_workspace.dummy
  ]
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_tags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["good"]
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["good"]
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_workspace_ids" "good" {
  tag_names    = ["good"]
  organization = tfe_workspace.foo.organization
  depends_on = [
    tfe_workspace.foo,
    tfe_workspace.bar,
    tfe_workspace.dummy
  ]
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_empty(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_workspace_ids" "foobar" {
  organization = tfe_workspace.foo.organization
}`, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_searchByTagAndName(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["bar"]
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["bar"]
}

data "tfe_workspace_ids" "good" {
  names        = ["workspace-foo-%d"]
  tag_names    = ["bar"]
  organization = tfe_workspace.foo.organization
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_namesEmpty(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["bar"]
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["bar"]
}

data "tfe_workspace_ids" "good" {
  names        = ["", "workspace-foo-%d"]
  tag_names    = ["bar"]
  organization = tfe_workspace.foo.organization
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_excludeTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["good", "happy"]
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["good"]
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = tfe_organization.foobar.id
}

data "tfe_workspace_ids" "good" {
  tag_names    = ["good"]
	exclude_tags    = ["happy"]
  organization = tfe_workspace.foo.organization
  depends_on = [
    tfe_workspace.foo,
    tfe_workspace.bar,
    tfe_workspace.dummy
  ]
}`, rInt, rInt, rInt, rInt)
}

func testAccTFEWorkspaceIDsDataSourceConfig_sameTagInTagNamesAndExcludeTags(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
  name         = "workspace-foo-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["good", "happy"]
}

resource "tfe_workspace" "bar" {
  name         = "workspace-bar-%d"
  organization = tfe_organization.foobar.id
  tag_names    = ["happy", "play"]
}

resource "tfe_workspace" "dummy" {
  name         = "workspace-dummy-%d"
  organization = tfe_organization.foobar.id
	tag_names    = ["good", "play", "happy"]
}

data "tfe_workspace_ids" "good" {
  tag_names    = ["good", "happy"]
	exclude_tags    = ["happy"]
  organization = tfe_workspace.foo.organization
  depends_on = [
    tfe_workspace.foo,
    tfe_workspace.bar,
    tfe_workspace.dummy
  ]
}`, rInt, rInt, rInt, rInt)
}
