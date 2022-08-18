package tfe

import (
	"os"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
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

// skips a test if the test requires a paid feature, and this flag
// SKIP_PAID is set.
func skipIfFreeOnly(t *testing.T) {
	skip := os.Getenv("SKIP_PAID") == "1"
	if skip {
		t.Skip("Skipping test that requires a paid feature. Remove 'SKIP_PAID=1' if you want to run this test")
	}
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

type retryableFn func() (interface{}, error)

func retry(maxRetries, secondsBetween int, f retryableFn) (interface{}, error) {
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
