package provider

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEQueryRun_WithExplicitConfigVersion(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-org-%d", rInt)),
		Email: tfe.String("admin@terraformer.inc"),
	})
	t.Cleanup(orgCleanup)

	parentWorkspace, _ := setupWorkspacesWithConfig(t, tfeClient, rInt, organization.Name, "test-fixtures/query-config")

	cvID := getLatestConfigurationVersionID(t, tfeClient, parentWorkspace.ID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEQueryRun_ExplicitConfig(parentWorkspace.ID, cvID),
				PostApplyFunc: func() {
					checkTFEQueryRunCreated(t, tfeClient, parentWorkspace.ID)
				},
			},
		},
	})
}

func TestAccTFEQueryRun_WaitForLatestConfiguration(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-org-%d", rInt)),
		Email: tfe.String("admin@terraformer.inc"),
	})
	t.Cleanup(orgCleanup)

	parentWorkspace, _ := setupWorkspacesWithConfig(t, tfeClient, rInt, organization.Name, "test-fixtures/query-config")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEQueryRun_WaitForLatest(parentWorkspace.ID),
				PostApplyFunc: func() {
					checkTFEQueryRunCreated(t, tfeClient, parentWorkspace.ID)
				},
			},
		},
	})
}

// Test 3: Validation Errors
func TestAccTFEQueryRun_ValidationErrors(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	tfeClient, err := getClientUsingEnv()
	if err != nil {
		t.Fatal(err)
	}

	organization, orgCleanup := createOrganization(t, tfeClient, tfe.OrganizationCreateOptions{
		Name:  tfe.String(fmt.Sprintf("tst-org-%d", rInt)),
		Email: tfe.String("admin@terraformer.inc"),
	})
	t.Cleanup(orgCleanup)

	parentWorkspace, _ := setupWorkspacesWithConfig(t, tfeClient, rInt, organization.Name, "test-fixtures/query-config")

	invalidCases := []struct {
		Config      string
		ExpectError *regexp.Regexp
	}{
		{
			Config:      testAccTFEQueryRun_MissingParams(parentWorkspace.ID),
			ExpectError: regexp.MustCompile("Configuration Version ID is required"),
		},
	}

	for _, invalidCase := range invalidCases {
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
			},
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			ProtoV6ProviderFactories: testAccMuxedProviders,
			Steps: []resource.TestStep{
				{
					Config:      invalidCase.Config,
					ExpectError: invalidCase.ExpectError,
				},
			},
		})
	}
}

func getLatestConfigurationVersionID(t *testing.T, client *tfe.Client, workspaceID string) string {
	listOpts := &tfe.ConfigurationVersionListOptions{
		ListOptions: tfe.ListOptions{PageSize: 1},
	}
	cvList, err := client.ConfigurationVersions.List(context.Background(), workspaceID, listOpts)
	if err != nil {
		t.Fatalf("Failed to list configuration versions: %s", err)
	}
	if len(cvList.Items) == 0 {
		t.Fatalf("No configuration versions found for workspace %s", workspaceID)
	}
	return cvList.Items[0].ID
}

func checkTFEQueryRunCreated(t *testing.T, client *tfe.Client, workspaceID string) {
	listOpts := &tfe.QueryRunListOptions{
		ListOptions: tfe.ListOptions{PageSize: 1},
	}

	err := retry.RetryContext(context.Background(), 30*time.Second, func() *retry.RetryError {
		runs, err := client.QueryRuns.List(context.Background(), workspaceID, listOpts)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error listing query runs: %w", err))
		}

		if len(runs.Items) == 0 {
			return retry.RetryableError(fmt.Errorf("No query runs found yet"))
		}

		// Ideally, we check that the run status is Finished, as the action waits for it.
		latestRun := runs.Items[0]
		if latestRun.Status != tfe.QueryRunFinished {
			return retry.NonRetryableError(fmt.Errorf("Expected query run to be finished, got %s", latestRun.Status))
		}

		return nil
	})

	if err != nil {
		t.Fatalf("PostApplyFunc assertion failed: %s", err)
	}
}

func testAccTFEQueryRun_ExplicitConfig(workspaceID, cvID string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_query_run.test]
    }
  }
}

action "tfe_query_run" "test" {
  config {
    workspace_id = "%s"
    configuration_version_id = "%s"
    variables = {
      "animals" = "5"
    }
  }
}`, workspaceID, cvID)
}

func testAccTFEQueryRun_WaitForLatest(workspaceID string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_query_run.test]
    }
  }
}

action "tfe_query_run" "test" {
  config {
    workspace_id = "%s"
    wait_for_latest_configuration = true
  }
}`, workspaceID)
}

func testAccTFEQueryRun_MissingParams(workspaceID string) string {
	return fmt.Sprintf(`
resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.tfe_query_run.test]
    }
  }
}

action "tfe_query_run" "test" {
  config {
    workspace_id = "%s"
    # Missing configuration_version_id AND wait_for_latest_configuration
  }
}`, workspaceID)
}
