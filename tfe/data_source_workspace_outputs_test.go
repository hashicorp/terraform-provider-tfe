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

func TestAccTFEWorkspaceOutputs(t *testing.T) {
	skipIfFreeOnly(t)

	client, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform.tfstate"
	orgName, wsName, orgCleanup, wsCleanup := createOutputs(t, client, rInt, fileName)
	defer orgCleanup()
	defer wsCleanup()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceOutputs_dataSource(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_outputs.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_outputs.foobar", "workspace", wsName),
					// This is the value in test-fixtures/state-versions/terraform.tfstate
					testCheckOutputState("identifier_output", &terraform.OutputState{
						Value: "9023256633839603543",
					}),
					// This is the value in test-fixtures/state-versions/terraform.tfstate
					testCheckOutputState("records_output", &terraform.OutputState{
						Value: []interface{}{"hashicorp.com", "terraform.io"},
					}),
				),
			},
		},
	})
}

func TestAccTFEWorkspaceOutputs_emptyOutputs(t *testing.T) {
	skipIfFreeOnly(t)

	client, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform-empty-outputs.tfstate"
	orgName, wsName, orgCleanup, wsCleanup := createOutputs(t, client, rInt, fileName)
	defer orgCleanup()
	defer wsCleanup()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceOutputs_dataSource_emptyOutputs(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_outputs.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_workspace_outputs.foobar", "workspace", wsName),
					// This is relies on test-fixtures/state-versions/terraform-empty-outputs.tfstate
					testCheckOutputState("state_output", &terraform.OutputState{
						Value: map[string]interface{}{},
					}),
				),
			},
		},
	})
}
func testCheckOutputState(name string, expectedOutputState *terraform.OutputState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.String() != expectedOutputState.String() {
			return fmt.Errorf("Expected the output state %s to match expected output state %s", rs.String(), expectedOutputState.String())
		}
		return nil
	}
}

func createOutputs(t *testing.T, client *tfe.Client, rInt int, fileName string) (string, string, func(), func()) {
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

	state, err := ioutil.ReadFile(fileName)
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

func testAccTFEWorkspaceOutputs_dataSource(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_workspace_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

output "identifier_output" {
	// the 'identifier' value references the value in outputs in
	// the file 'test-fixtures/state-versions/terraform.tfstate
	value = data.tfe_workspace_outputs.foobar.values.identifier
}
output "records_output" {
	// the 'records' value references the value in outputs in
	// the file 'test-fixtures/state-versions/terraform.tfstate
	value = data.tfe_workspace_outputs.foobar.values.records
}`, rInt, rInt, org, workspace)
}

func testAccTFEWorkspaceOutputs_dataSource_emptyOutputs(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_workspace_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

output "state_output" {
	// this relies on the file 'test-fixtures/state-versions/terraform-empty-outputs.tfstate
	value = data.tfe_workspace_outputs.foobar.values
}`, rInt, rInt, org, workspace)
}
