// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOutputs(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	waitForOutputs(t, tfeClient, orgName, wsName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOutputs_dataSource(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "workspace", wsName),
					// These outputs rely on the values in test-fixtures/state-versions/terraform.tfstate
					testCheckOutputState("test_output_list_string", &terraform.OutputState{Value: []interface{}{"us-west-1a"}}),
					testCheckOutputState("test_output_string", &terraform.OutputState{Value: "9023256633839603543"}),
					testCheckOutputState("test_output_tuple_number", &terraform.OutputState{Value: []interface{}{"1", "2"}}),
					testCheckOutputState("test_output_tuple_string", &terraform.OutputState{Value: []interface{}{"one", "two"}}),
					testCheckOutputState("test_output_object", &terraform.OutputState{Value: map[string]interface{}{"foo": "bar"}}),
					testCheckOutputState("test_output_number", &terraform.OutputState{Value: "5"}),
					testCheckOutputState("test_output_bool", &terraform.OutputState{Value: "true"}),
				),
			},
		},
	})
}

func TestAccTFEOutputs_ReadAllNonSensitiveValues(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	waitForOutputs(t, tfeClient, orgName, wsName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOutputs_dataSourceReadNonsensitiveValues(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "workspace", wsName),
					// nonsensitive_values does not set sensitive values
					resource.TestCheckNoResourceAttr("data.tfe_outputs.foobar", "nonsensitive_values.test_output_string"),
					// These outputs rely on the values in test-fixtures/state-versions/terraform.tfstate
					testCheckOutputState("test_output_list_string", &terraform.OutputState{Value: []interface{}{"us-west-1a"}}),
					testCheckOutputState("test_output_tuple_number", &terraform.OutputState{Value: []interface{}{"1", "2"}}),
					testCheckOutputState("test_output_tuple_string", &terraform.OutputState{Value: []interface{}{"one", "two"}}),
					testCheckOutputState("test_output_object", &terraform.OutputState{Value: map[string]interface{}{"foo": "bar"}}),
					testCheckOutputState("test_output_number", &terraform.OutputState{Value: "5"}),
					testCheckOutputState("test_output_bool", &terraform.OutputState{Value: "true"}),
				),
			},
		},
	})
}

func TestAccTFEOutputs_emptyOutputs(t *testing.T) {
	skipIfUnitTest(t)

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatalf("error getting client %v", err)
	}

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	fileName := "test-fixtures/state-versions/terraform-empty-outputs.tfstate"
	orgName, wsName, orgCleanup := createStateVersion(t, tfeClient, rInt, fileName)
	t.Cleanup(orgCleanup)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOutputs_dataSource_emptyOutputs(rInt, orgName, wsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"tfe_organization.foobar", "name", fmt.Sprintf("tst-%d", rInt)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", fmt.Sprintf("workspace-test-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "organization", orgName),
					resource.TestCheckResourceAttr(
						"data.tfe_outputs.foobar", "workspace", wsName),
					// This relies on test-fixtures/state-versions/terraform-empty-outputs.tfstate
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
			return fmt.Errorf("not found: %s", name)
		}
		if rs.String() != expectedOutputState.String() {
			return fmt.Errorf("expected the output state %s to match expected output state %s", rs.String(), expectedOutputState.String())
		}
		return nil
	}
}

func createStateVersion(t *testing.T, client *tfe.Client, rInt int, fileName string) (string, string, func()) {
	t.Helper()
	var orgCleanup func()

	org, err := client.Organizations.Create(ctx, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-terraform-%d", rInt)),
		Email: tfe.String(fmt.Sprintf("%d@tfe.local", rInt)),
	})
	if err != nil {
		t.Fatal(err)
	}

	upgradeOrganizationSubscription(t, client, org)

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

	state, err := os.ReadFile(fileName)
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

	return org.Name, ws.Name, orgCleanup
}

func waitForOutputs(t *testing.T, client *tfe.Client, org, workspace string) {
	t.Helper()
	ws, err := client.Workspaces.Read(ctx, org, workspace)

	if err != nil {
		t.Fatal(err)
	}

	maxRetries := 15
	secondsToWait := 4

	// Wait for outputs to be populated
	_, err = retryFn(maxRetries, secondsToWait, func() (interface{}, error) {
		svo, oerr := client.StateVersionOutputs.ReadCurrent(ctx, ws.ID)
		if oerr != nil {
			return nil, fmt.Errorf("could not read outputs: %w", oerr)
		}

		if len(svo.Items) == 0 {
			return nil, errors.New("outputs are not ready")
		}

		return svo.Items, nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func testAccTFEOutputs_dataSource(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

// All of these values reference the outputs in the file
// 'test-fixtures/state-versions/terraform.tfstate
output "test_output_list_string" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_list_string
}
output "test_output_string" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_string
}
output "test_output_tuple_number" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_tuple_number
}
output "test_output_tuple_string" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_tuple_string
}
output "test_output_object" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_object
}
output "test_output_number" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_number
}
output "test_output_bool" {
	sensitive = true
	value = data.tfe_outputs.foobar.values.test_output_bool
}
`, rInt, rInt, org, workspace)
}

func testAccTFEOutputs_dataSourceReadNonsensitiveValues(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-%d"
	email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
	name                  = "workspace-test-%d"
	organization          = tfe_organization.foobar.name
}

data "tfe_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

// All of these values reference the outputs in the file
// 'test-fixtures/state-versions/terraform.tfstate except the sensitive attr test_output_string
output "test_output_list_string" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_list_string
}
output "test_output_tuple_number" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_tuple_number
}
output "test_output_tuple_string" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_tuple_string
}
output "test_output_object" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_object
}
output "test_output_number" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_number
}
output "test_output_bool" {
	value = data.tfe_outputs.foobar.nonsensitive_values.test_output_bool
}
`, rInt, rInt, org, workspace)
}

func testAccTFEOutputs_dataSource_emptyOutputs(rInt int, org, workspace string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test-%d"
  organization          = tfe_organization.foobar.name
}

data "tfe_outputs" "foobar" {
  organization = "%s"
  workspace = "%s"
}

output "state_output" {
	// this relies on the file 'test-fixtures/state-versions/terraform-empty-outputs.tfstate
	value = nonsensitive(data.tfe_outputs.foobar.values)
}`, rInt, rInt, org, workspace)
}
