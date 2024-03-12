// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const RunTasksURLEnvName = "RUN_TASKS_URL"

type testClientOptions struct {
	defaultOrganization          string
	defaultWorkspaceID           string
	remoteStateConsumersResponse string
}

// testTfeClient creates a mock client that creates workspaces with their ID
// set to workspaceID.
func testTfeClient(t *testing.T, options testClientOptions) *tfe.Client {
	config := &tfe.Config{
		Token: "not-a-token",
	}

	if options.defaultOrganization == "" {
		options.defaultOrganization = "hashicorp"
	}
	if options.defaultWorkspaceID == "" {
		options.defaultWorkspaceID = "ws-testing"
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		t.Fatalf("error creating tfe client: %v", err)
	}

	client.Workspaces = newMockWorkspaces(options)

	return client
}

// Attempts to upgrade an organization to the business plan. Requires a user token with admin access.
// DEPRECATED : Please use the newSubscriptionUpdater instead.
func upgradeOrganizationSubscription(t *testing.T, _ *tfe.Client, organization *tfe.Organization) {
	newSubscriptionUpdater(organization).WithBusinessPlan().Update(t)
}

func createBusinessOrganization(t *testing.T, client *tfe.Client) (*tfe.Organization, func()) {
	org, orgCleanup := createOrganization(t, client, tfe.OrganizationCreateOptions{
		Name:  tfe.String("tst-" + randomString(t)),
		Email: tfe.String(fmt.Sprintf("%s@tfe.local", randomString(t))),
	})

	newSubscriptionUpdater(org).WithBusinessPlan().Update(t)

	return org, orgCleanup
}

func createOrganization(t *testing.T, client *tfe.Client, options tfe.OrganizationCreateOptions) (*tfe.Organization, func()) {
	ctx := context.Background()
	org, err := client.Organizations.Create(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	return org, func() {
		if err := client.Organizations.Delete(ctx, org.Name); err != nil {
			t.Errorf("Error destroying organization! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Organization:%s\nError: %s", org.Name, err)
		}
	}
}

func createTempWorkspace(t *testing.T, client *tfe.Client, orgName string) *tfe.Workspace {
	t.Helper()

	ctx := context.Background()
	ws, err := client.Workspaces.Create(ctx, orgName, tfe.WorkspaceCreateOptions{
		Name: tfe.String(fmt.Sprintf("tst-workspace-%s", randomString(t))),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := client.Workspaces.DeleteByID(ctx, ws.ID); err != nil {
			t.Errorf("Error destroying workspace! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Workspace:%s\nError: %s", ws.ID, err)
		}
	})

	return ws
}

func createOrganizationMembership(t *testing.T, client *tfe.Client, orgName string, options tfe.OrganizationMembershipCreateOptions) *tfe.OrganizationMembership {
	ctx := context.Background()
	orgMembership, err := client.OrganizationMemberships.Create(ctx, orgName, options)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := client.OrganizationMemberships.Delete(ctx, orgMembership.ID); err != nil {
			t.Errorf("Error destroying organization membership! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Organization memberships:%s\nError: %s", orgMembership.ID, err)
		}
	})

	return orgMembership
}

func createAndUploadConfigurationVersion(t *testing.T, workspace *tfe.Workspace, tfeClient *tfe.Client, configPath string) *tfe.ConfigurationVersion {
	cv, err := tfeClient.ConfigurationVersions.Create(ctx, workspace.ID, tfe.ConfigurationVersionCreateOptions{AutoQueueRuns: tfe.Bool(false)})
	if err != nil {
		t.Fatal(err)
	}

	err = tfeClient.ConfigurationVersions.Upload(ctx, cv.UploadURL, configPath)
	if err != nil {
		t.Fatal(err)
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Context canceled while waiting for the configuration version to be uploaded: %w", ctx.Err())
		case <-ticker.C:
			cv, err = tfeClient.ConfigurationVersions.Read(ctx, cv.ID)
			if err != nil {
				t.Fatal(err)
			}

			switch cv.Status {
			case tfe.ConfigurationUploaded:
				return cv
			case tfe.ConfigurationFetching, tfe.ConfigurationPending:
				t.Logf("Waiting for the configuration version to be uploaded for workspace %s...", workspace.ID)
			default:
				t.Fatalf("Configuration version entered unexpected state %s", cv.Status)
			}
		}
	}
}

func createProject(t *testing.T, client *tfe.Client, orgName string, options tfe.ProjectCreateOptions) *tfe.Project {
	ctx := context.Background()
	proj, err := client.Projects.Create(ctx, orgName, options)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := client.Projects.Delete(ctx, proj.ID); err != nil {
			t.Errorf("Error destroying project! WARNING: Dangling resources\n"+
				"may exist! The full error is show below:\n\n"+
				"Project:%s\nError: %s", proj.ID, err)
		}
	})

	return proj
}

func skipIfCloud(t *testing.T) {
	if !enterpriseEnabled() {
		t.Skip("Skipping test for a feature unavailable in Terraform Cloud. Set 'ENABLE_TFE=1' to run.")
	}
}

func skipIfEnterprise(t *testing.T) {
	if enterpriseEnabled() {
		t.Skip("Skipping test for a feature unavailable in Terraform Enterprise. Set 'ENABLE_TFE=0' to run.")
	}
}

func skipUnlessRunTasksDefined(t *testing.T) {
	if value, ok := os.LookupEnv(RunTasksURLEnvName); !ok || value == "" {
		t.Skipf("Skipping tests for Run Tasks. Set '%s' to enabled this tests.", RunTasksURLEnvName)
	}
}

func skipUnlessBeta(t *testing.T) {
	if !betaFeaturesEnabled() {
		t.Skip("Skipping test related to a Terraform Cloud/Enterprise beta feature. Set ENABLE_BETA=1 to run.")
	}
}

func enterpriseEnabled() bool {
	return os.Getenv("ENABLE_TFE") == "1"
}

func isAcceptanceTest() bool {
	return os.Getenv("TF_ACC") == "1"
}

func runTasksURL() string {
	return os.Getenv(RunTasksURLEnvName)
}

// Checks to see if ENABLE_BETA is set to 1, thereby enabling tests for beta features.
func betaFeaturesEnabled() bool {
	return os.Getenv("ENABLE_BETA") == "1"
}

// Most tests rely on terraform-plugin-sdk/helper/resource.Test to run.  That test helper ensures
// that TF_ACC=1 or else it skips. In some rare cases, however, tests do not use the SDK helper and
// are acceptance tests.
// This `skipIfUnitTest` is used when you are doing some extra setup work that may fail when `go
// test` is run without additional environment variables for acceptance tests. By adding this at the
// top of the test, it will skip the test if `TF_ACC=1` is not set.
func skipIfUnitTest(t *testing.T) {
	if !isAcceptanceTest() {
		t.Skip("Skipping test because this test is an acceptance test, and is run as a unit test. Set 'TF_ACC=1' to run.")
	}
}

// A wrapper for resource.TestCheckResourceAttr that skips the check if running tests against
// Terraform Enterprise. Useful for testing new attributes that haven't been added to TFE
// yet, without having to skip an entire test.
//
//nolint:unparam
func testCheckResourceAttrUnlessEnterprise(name, key, value string) resource.TestCheckFunc {
	if enterpriseEnabled() {
		return func(s *terraform.State) error {
			return nil
		}
	}
	return resource.TestCheckResourceAttr(name, key, value)
}

func randomString(t *testing.T) string {
	v, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

type retryableFn func() (interface{}, error)

func retryFn(maxRetries, secondsBetween int, f retryableFn) (interface{}, error) {
	tick := time.NewTicker(time.Duration(secondsBetween) * time.Second)
	retries := 0

	defer tick.Stop()

	for {
		<-tick.C
		res, err := f()
		if err == nil {
			return res, nil
		}

		if retries >= maxRetries {
			return nil, err
		}

		retries += 1
	}
}
