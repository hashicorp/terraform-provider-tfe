// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTFEOrgMaxTokenTTLPolicyDataSource_basic(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_basic(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.test", "organization", orgName),
					// Check millisecond attributes are set
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "org_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "team_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "audit_trail_token_max_ttl_ms"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "user_token_max_ttl_ms"),
					// Check human-readable duration attributes are set
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "org_token_max_ttl"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "team_token_max_ttl"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "audit_trail_token_max_ttl"),
					resource.TestCheckResourceAttrSet("data.tfe_org_max_token_ttl_policy.test", "user_token_max_ttl"),
				),
			},
		},
	})
}

func TestAccTFEOrgMaxTokenTTLPolicyDataSource_withResource(t *testing.T) {
	skipUnlessBeta(t)
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_withResource(orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check organization has max_ttl_enabled
					resource.TestCheckResourceAttr("tfe_organization.foo", "max_ttl_enabled", "true"),
					// Check resource
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.foo", "org_token_max_ttl", "30d"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.foo", "team_token_max_ttl", "7d"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.foo", "user_token_max_ttl", "5y"),
					resource.TestCheckResourceAttr("tfe_org_max_token_ttl_policy.foo", "audit_trail_token_max_ttl", "90d"),
					// Check data source returns milliseconds (API values)
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.foo", "org_token_max_ttl_ms", "2592000000"),         // 30 days
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.foo", "team_token_max_ttl_ms", "604800000"),         // 7 days
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.foo", "user_token_max_ttl_ms", "157680000000"),      // 5 years
					resource.TestCheckResourceAttr("data.tfe_org_max_token_ttl_policy.foo", "audit_trail_token_max_ttl_ms", "7776000000"), // 90 days
				),
			},
		},
	})
}

func TestFetchTokenTTLPoliciesV2_largeMaxTTL(t *testing.T) {
	orgName := "hashicorp"

	// 5 years in milliseconds, matching the acceptance test's "5y" case.
	// This exceeds math.MaxInt32 (2147483647) and would previously fail to
	// deserialize when max-ttl-ms was typed as *int32.
	const largeMaxTTLMs = int64(157680000000)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/organizations/"+orgName+"/token-ttl-policies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		fmt.Fprintf(w, `{"data": [
			{"id": "ttl-1", "type": "organization-token-ttl-policies", "attributes": {"token-type": "organization", "max-ttl-ms": %d}},
			{"id": "ttl-2", "type": "organization-token-ttl-policies", "attributes": {"token-type": "user", "max-ttl-ms": %d}}
		]}`, int64(2592000000), largeMaxTTLMs)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"status":"404","title":"not found"}]}`, http.StatusNotFound)
	})

	client := testTfeClientV2(t, mux)

	resp, err := client.API.Organizations().ByOrganization_name(orgName).TokenTtlPolicies().Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error deserializing a max-ttl-ms beyond int32 range, got: %v", err)
	}

	result := modelFromTokenTTLPoliciesData(orgName, resp.GetData())

	if got := result.UserTokenMaxTTLMs.ValueInt64(); got != largeMaxTTLMs {
		t.Fatalf("wrong user_token_max_ttl_ms\ngot: %d\nwant: %d", got, largeMaxTTLMs)
	}
	if got := result.OrgTokenMaxTTLMs.ValueInt64(); got != 2592000000 {
		t.Fatalf("wrong org_token_max_ttl_ms\ngot: %d\nwant: %d", got, 2592000000)
	}
}

func testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_basic(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "test" {
  name  = "%s"
  email = "admin@company.com"
}

data "tfe_org_max_token_ttl_policy" "test" {
  organization = tfe_organization.test.name
}
`, orgName)
}

func testAccTFEOrgMaxTokenTTLPolicyDataSourceConfig_withResource(orgName string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foo" {
  name             = "%s"
  email            = "admin@company.com"
  max_ttl_enabled  = true
}

resource "tfe_org_max_token_ttl_policy" "foo" {
  organization              = tfe_organization.foo.name
  org_token_max_ttl         = "30d"
  team_token_max_ttl        = "7d"
  user_token_max_ttl        = "5y"
  audit_trail_token_max_ttl = "90d"
}

data "tfe_org_max_token_ttl_policy" "foo" {
  organization = tfe_org_max_token_ttl_policy.foo.organization
  depends_on   = [tfe_org_max_token_ttl_policy.foo]
}
`, orgName)
}
