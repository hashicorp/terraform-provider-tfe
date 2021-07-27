package tfe

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEStateOutputs(t *testing.T) {
	skipIfFreeOnly(t)

	client, err := getClientByEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	orgName, wsName, orgCleanup, wsCleanup := createOutputs(t, client, rInt)
	defer orgCleanup()
	defer wsCleanup()

	expectedOutput := "hello world"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEStateOutputs_dataSource(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFooBar(orgName),
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckOutput(
						"states", expectedOutput),
				),
			},
		},
	})
}
func testFooBar(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("OMAR WE OUT HERE")
		fmt.Println(n)
		time.Sleep(5 * time.Second)
		return nil
	}
}

func createOutputs(t *testing.T, client *tfe.Client, rInt int) (string, string, func(), func()) {
	var orgCleanup func()
	var wsCleanup func()

	org, err := client.Organizations.Create(ctx, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%d@tfe.local", rInt)),
	})
	if err != nil {
		t.Fatal(err)
	}
	orgCleanup = func() {
		if err := client.Organizations.Delete(ctx, org.Name); err != nil {
			t.Errorf("Error destroying organization! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Organization: %s\nError: %s", org.Name, err)
		}
	}

	ws, err := client.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
		Name: tfe.String(fmt.Sprintf("tst-workspace-test-%d", rInt)),
	})
	if err != nil {
		t.Fatal(err)
	}
	wsCleanup = func() {
		if err := client.Workspaces.Delete(ctx, org.Name, ws.Name); err != nil {
			t.Errorf("Error destroying workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is shown below.\n\n"+
				"Workspace: %s\nError: %s", ws.Name, err)
		}
	}

	state, err := ioutil.ReadFile("test-fixtures/state-versions/terraform.tfstate")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workspaces.Lock(ctx, ws.ID, tfe.WorkspaceLockOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, err := client.Workspaces.Unlock(ctx, ws.ID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	_, err = client.StateVersions.Create(ctx, ws.ID, tfe.StateVersionCreateOptions{
		MD5:    tfe.String(fmt.Sprintf("%x", md5.Sum(state))),
		Serial: tfe.Int64(0),
		State:  tfe.String(base64.StdEncoding.EncodeToString(state)),
	})
	if err != nil {
		t.Fatal(err)
	}

	return org.Name, ws.Name, orgCleanup, wsCleanup
}

func testAccTFEStateOutputs_defaultOutputs(rInt int, output string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

output "foo" {
	value = "%s"
}`, rInt, rInt, output)
}

func testAccTFEStateOutputs_dataSource(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_state_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

output "states" {
	// this references the 'output "foo"' in the testAccTFEStateOutputs_defaultOutputs config
 // value = data.tfe_state_outputs.foobar.values.foo
	value = "hello world"
}`, rInt, rInt, org, workspace)
}
