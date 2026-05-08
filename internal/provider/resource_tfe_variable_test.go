// Copyright IBM Corp. 2018, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-version"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTFEVariable_basic(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariable_basic_variable_set(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic_variable_set(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableiSetVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariable_update(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributes(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "some description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "env"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "false"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},

			{
				Config: testAccTFEVariable_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					testAccCheckTFEVariableAttributesUpdate(variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

func TestAccTFEVariable_update_key_sensitive(t *testing.T) {
	first := &tfe.Variable{}
	second := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", first),
					testAccCheckTFEVariableAttributesUpdate(first),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "true"),
				),
			},
			{
				Config: testAccTFEVariable_update_key_sensitive(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", second),
					testAccCheckTFEVariableAttributesUpdate_key_sensitive(second),
					testAccCheckTFEVariableIDsNotEqual(first, second),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "key", "key_updated_2"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_updated"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "description", "another description"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "category", "terraform"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "hcl", "true"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "true"),
				),
			},
		},
	})
}

func TestAccTFEVariable_valueWriteOnly(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42
	versionOne, versionTwo := 1, 2

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEVariable_valueAndValueWO(rInt, variableValue1, false),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo" is specified`),
			},
			{
				Config:      testAccTFEVariable_valueWOOnly(rInt, variableValue1),
				ExpectError: regexp.MustCompile(`Attribute "value_wo_version" must be specified when "value_wo" is specified`),
			},
			{
				Config:      testAccTFEVariable_versionOnly(rInt),
				ExpectError: regexp.MustCompile(`Attribute "value_wo" must be specified when "value_wo_version" is specified`),
			},
			{
				Config:      testAccTFEVariable_valueWithVersion(rInt),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo_version" is specified`),
			},
			{
				Config: testAccTFEVariable_valueWriteOnly(rInt, variableValue1, versionOne, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value_wo_version", fmt.Sprintf("%d", versionOne)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
			{
				Config: testAccTFEVariable_valueWriteOnly(rInt, variableValue2, versionTwo, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value_wo_version", fmt.Sprintf("%d", versionTwo)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariable_valueWriteOnly_variable_set(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		ProtoV6ProviderFactories: testAccMuxedProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTFEVariable_valueAndValueWO_variable_set(rInt, variableValue1, false),
				ExpectError: regexp.MustCompile(`Attribute "value" cannot be specified when "value_wo" is specified`),
			},
			{
				Config: testAccTFEVariable_valueWriteOnly_variable_set(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
			{
				Config: testAccTFEVariable_valueWriteOnly_variable_set(rInt, variableValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
				),
			},
		},
	})
}

func TestAccTFEVariable_updateValueWriteOnlyToValue(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	versionOne := 1

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_valueWriteOnly(rInt, variableValue, versionOne, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", ""),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value_wo_version", "1"),
				),
			},
			{
				Config: testAccTFEVariable_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", "value_test"),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo"),
					resource.TestCheckNoResourceAttr(
						"tfe_variable.foobar", "value_wo_version"),
				),
			},
		},
	})
}

func TestAccTFEVariable_readable_value(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that downstream resources may depend on both the value and readableValue of a non-sensitive variable
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue2)),
				),
			},
		},
	})
}

func TestAccTFEVariable_readable_value_becomes_sensitive(t *testing.T) {
	variable := &tfe.Variable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that if an insensitive variable becomes sensitive, downstream resources depending on the readableValue
	// may no longer access it, but that the value may still be used directly
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config:      testAccTFEVariable_readablevalue(rInt, variableValue2, true),
				ExpectError: regexp.MustCompile(`tfe_variable.foobar.readable_value is null`),
			},
		},
	})
}

func TestAccTFEVariable_varset_readable_value(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that downstream resources may depend on both the value and readableValue of a non-sensitive variable
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue2)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue2)),
				),
			},
		},
	})
}

func TestAccTFEVariable_varset_readable_value_becomes_sensitive(t *testing.T) {
	variable := &tfe.VariableSetVariable{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	variableValue1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	variableValue2 := variableValue1 + 42

	// Test that if an insensitive variable becomes sensitive, downstream resources depending on the readableValue
	// may no longer access it, but that the value may still be used directly
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_varset_readablevalue(rInt, variableValue1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEVariableSetVariableExists(
						"tfe_variable.foobar", variable),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "value", fmt.Sprintf("%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_variable.foobar", "sensitive", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-readable", "name", fmt.Sprintf("downstream-readable-%d", variableValue1)),
					resource.TestCheckResourceAttr(
						"tfe_workspace.downstream-nonreadable", "name", fmt.Sprintf("downstream-nonreadable-%d", variableValue1)),
				),
			},
			{
				Config:      testAccTFEVariable_varset_readablevalue(rInt, variableValue2, true),
				ExpectError: regexp.MustCompile(`tfe_variable.foobar.readable_value is null`),
			},
		},
	})
}

func TestAccTFEVariable_importIdentityWithWorkspace(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_variable.foobar", map[string]knownvalue.Check{
						"id":              knownvalue.NotNull(),
						"configurable_id": knownvalue.NotNull(),
						"hostname":        knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},

			{
				ResourceName:    "tfe_variable.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccTFEVariable_importIdentityWithVarset(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic_variable_set(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_variable.foobar", map[string]knownvalue.Check{
						"id":              knownvalue.NotNull(),
						"configurable_id": knownvalue.NotNull(),
						"hostname":        knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},

			{
				ResourceName:    "tfe_variable.foobar",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccTFEVariable_import(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_basic(rInt),
			},

			{
				ResourceName:        "tfe_variable.foobar",
				ImportState:         true,
				ImportStateIdPrefix: fmt.Sprintf("tst-terraform-%d/workspace-test/", rInt),
				ImportStateVerify:   true,
			},
		},
	})
}

func TestAccTFEVariable_mutableIdentity(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccMuxedProviders,
		CheckDestroy:             testAccCheckTFEVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEVariable_mutableIdentity_workspace(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_variable.foobar", map[string]knownvalue.Check{
						"id":              knownvalue.NotNull(),
						"configurable_id": knownvalue.StringRegexp(regexp.MustCompile(`^ws-.*$`)),
						"hostname":        knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},
			{
				Config: testAccTFEVariable_mutableIdentity_varset(rInt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity("tfe_variable.foobar", map[string]knownvalue.Check{
						"id":              knownvalue.NotNull(),
						"configurable_id": knownvalue.StringRegexp(regexp.MustCompile(`^varset-.*$`)),
						"hostname":        knownvalue.StringExact(os.Getenv("TFE_HOSTNAME")),
					}),
				},
			},
		},
	})
}

type notFoundVariables struct{}

func (notFoundVariables) List(_ context.Context, _ string, _ *tfe.VariableListOptions) (*tfe.VariableList, error) {
	return nil, nil
}

func (notFoundVariables) ListAll(_ context.Context, _ string, _ *tfe.VariableListOptions) (*tfe.VariableList, error) {
	return nil, nil
}

func (notFoundVariables) Create(_ context.Context, _ string, _ tfe.VariableCreateOptions) (*tfe.Variable, error) {
	return nil, nil
}

func (notFoundVariables) Read(_ context.Context, _ string, _ string) (*tfe.Variable, error) {
	return nil, tfe.ErrResourceNotFound
}

func (notFoundVariables) Update(_ context.Context, _ string, _ string, _ tfe.VariableUpdateOptions) (*tfe.Variable, error) {
	return nil, nil
}

func (notFoundVariables) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

type notFoundVariableSetVariables struct{}

func (notFoundVariableSetVariables) List(_ context.Context, _ string, _ *tfe.VariableSetVariableListOptions) (*tfe.VariableSetVariableList, error) {
	return nil, nil
}

func (notFoundVariableSetVariables) Create(_ context.Context, _ string, _ *tfe.VariableSetVariableCreateOptions) (*tfe.VariableSetVariable, error) {
	return nil, nil
}

func (notFoundVariableSetVariables) Read(_ context.Context, _ string, _ string) (*tfe.VariableSetVariable, error) {
	return nil, tfe.ErrResourceNotFound
}

func (notFoundVariableSetVariables) Update(_ context.Context, _ string, _ string, _ *tfe.VariableSetVariableUpdateOptions) (*tfe.VariableSetVariable, error) {
	return nil, nil
}

func (notFoundVariableSetVariables) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

func TestResourceTFEVariableRead_RemovedWorkspaceVariableBackfillsIdentity(t *testing.T) {
	ctx := context.Background()
	client := testTfeClient(t, testClientOptions{})
	client.Variables = notFoundVariables{}

	r := &resourceTFEVariable{config: ConfiguredClient{Client: client}}

	readResp := runRemovedVariableRead(t, ctx, r, modelTFEVariable{
		ID:             types.StringValue("var-123"),
		Key:            types.StringValue("key_test"),
		Value:          types.StringValue("value_test"),
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  types.StringValue("value_test"),
		Category:       types.StringValue(string(tfe.CategoryEnv)),
		Description:    types.StringValue(""),
		HCL:            types.BoolValue(false),
		Sensitive:      types.BoolValue(false),
		WorkspaceID:    types.StringValue("ws-123"),
		VariableSetID:  types.StringNull(),
	})

	assertRemovedVariableRead(t, ctx, readResp, modelTFEVariableIdentity{
		ID:             types.StringValue("var-123"),
		ConfigurableID: types.StringValue("ws-123"),
		Hostname:       types.StringValue(client.BaseURL().Host),
	})
}

func TestResourceTFEVariableRead_RemovedVariableSetVariableBackfillsIdentity(t *testing.T) {
	ctx := context.Background()
	client := testTfeClient(t, testClientOptions{})
	client.VariableSetVariables = notFoundVariableSetVariables{}

	r := &resourceTFEVariable{config: ConfiguredClient{Client: client}}

	readResp := runRemovedVariableRead(t, ctx, r, modelTFEVariable{
		ID:             types.StringValue("var-456"),
		Key:            types.StringValue("key_test"),
		Value:          types.StringValue("value_test"),
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  types.StringValue("value_test"),
		Category:       types.StringValue(string(tfe.CategoryEnv)),
		Description:    types.StringValue(""),
		HCL:            types.BoolValue(false),
		Sensitive:      types.BoolValue(false),
		WorkspaceID:    types.StringNull(),
		VariableSetID:  types.StringValue("varset-123"),
	})

	assertRemovedVariableRead(t, ctx, readResp, modelTFEVariableIdentity{
		ID:             types.StringValue("var-456"),
		ConfigurableID: types.StringValue("varset-123"),
		Hostname:       types.StringValue(client.BaseURL().Host),
	})
}

func TestResourceTFEVariableRead_RemovedWorkspaceVariablePreservesExistingIdentity(t *testing.T) {
	ctx := context.Background()
	client := testTfeClient(t, testClientOptions{})
	client.Variables = notFoundVariables{}

	r := &resourceTFEVariable{config: ConfiguredClient{Client: client}}
	existingIdentity := &modelTFEVariableIdentity{
		ID:             types.StringValue("var-existing"),
		ConfigurableID: types.StringValue("ws-existing"),
		Hostname:       types.StringValue("preserve.example.com"),
	}

	readResp := runRemovedVariableRead(t, ctx, r, modelTFEVariable{
		ID:             types.StringValue("var-123"),
		Key:            types.StringValue("key_test"),
		Value:          types.StringValue("value_test"),
		ValueWO:        types.StringNull(),
		ValueWOVersion: types.Int64Null(),
		ReadableValue:  types.StringValue("value_test"),
		Category:       types.StringValue(string(tfe.CategoryEnv)),
		Description:    types.StringValue(""),
		HCL:            types.BoolValue(false),
		Sensitive:      types.BoolValue(false),
		WorkspaceID:    types.StringValue("ws-123"),
		VariableSetID:  types.StringNull(),
	}, existingIdentity)

	assertRemovedVariableRead(t, ctx, readResp, *existingIdentity)
}

func runRemovedVariableRead(t *testing.T, ctx context.Context, r *resourceTFEVariable, stateData modelTFEVariable, existingIdentity ...*modelTFEVariableIdentity) fwresource.ReadResponse {
	t.Helper()

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &stateData); diags.HasError() {
		t.Fatalf("unexpected state set diagnostics: %v", diags)
	}

	identitySchemaResp := &fwresource.IdentitySchemaResponse{}
	r.IdentitySchema(ctx, fwresource.IdentitySchemaRequest{}, identitySchemaResp)
	nullIdentity := tftypes.NewValue(identitySchemaResp.IdentitySchema.Type().TerraformType(ctx), nil)

	requestIdentity := &tfsdk.ResourceIdentity{
		Schema: identitySchemaResp.IdentitySchema,
		Raw:    nullIdentity.Copy(),
	}
	responseIdentity := &tfsdk.ResourceIdentity{
		Schema: identitySchemaResp.IdentitySchema,
		Raw:    nullIdentity.Copy(),
	}

	if len(existingIdentity) > 0 && existingIdentity[0] != nil {
		if diags := requestIdentity.Set(ctx, existingIdentity[0]); diags.HasError() {
			t.Fatalf("unexpected request identity diagnostics: %v", diags)
		}
		if diags := responseIdentity.Set(ctx, existingIdentity[0]); diags.HasError() {
			t.Fatalf("unexpected response identity diagnostics: %v", diags)
		}
	}

	readResp := fwresource.ReadResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    state.Raw.Copy(),
		},
		Identity: responseIdentity,
	}

	r.Read(ctx, fwresource.ReadRequest{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    state.Raw.Copy(),
		},
		Identity: requestIdentity,
	}, &readResp)

	return readResp
}

func assertRemovedVariableRead(t *testing.T, ctx context.Context, readResp fwresource.ReadResponse, expectedIdentity modelTFEVariableIdentity) {
	t.Helper()

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected read diagnostics: %v", readResp.Diagnostics)
	}

	if !readResp.State.Raw.IsFullyNull() {
		t.Fatalf("expected resource to be removed from state, got %s", readResp.State.Raw.String())
	}

	if readResp.Identity == nil {
		t.Fatal("expected resource identity to be preserved")
	}

	if readResp.Identity.Raw.IsFullyNull() {
		t.Fatal("expected resource identity to be backfilled for removed resource")
	}

	var gotIdentity modelTFEVariableIdentity
	if diags := readResp.Identity.Get(ctx, &gotIdentity); diags.HasError() {
		t.Fatalf("unexpected identity diagnostics: %v", diags)
	}

	if gotIdentity.ID.ValueString() != expectedIdentity.ID.ValueString() {
		t.Fatalf("expected identity id %q, got %q", expectedIdentity.ID.ValueString(), gotIdentity.ID.ValueString())
	}

	if gotIdentity.ConfigurableID.ValueString() != expectedIdentity.ConfigurableID.ValueString() {
		t.Fatalf("expected configurable_id %q, got %q", expectedIdentity.ConfigurableID.ValueString(), gotIdentity.ConfigurableID.ValueString())
	}

	if gotIdentity.Hostname.ValueString() != expectedIdentity.Hostname.ValueString() {
		t.Fatalf("expected hostname %q, got %q", expectedIdentity.Hostname.ValueString(), gotIdentity.Hostname.ValueString())
	}
}

// Verify that the rewritten framework version of the resource results in no
// changes when upgrading from the final sdk v2 version of the resource.
func TestAccTFEVariable_rewrite(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"tfe": {
						VersionConstraint: "0.44.1",
						Source:            "hashicorp/tfe",
					},
				},
				Config: testAccTFEVariable_everything(rInt),
				// leaving Check empty, we just care that they're the same
			},
			{
				ProtoV6ProviderFactories: testAccMuxedProviders,
				Config:                   testAccTFEVariable_everything(rInt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckTFEVariableExists(
	n string, variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		wsID := rs.Primary.Attributes["workspace_id"]
		ws, err := testAccConfiguredClient.Client.Workspaces.ReadByID(ctx, wsID)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving workspace %s: %w", wsID, err)
		}

		v, err := testAccConfiguredClient.Client.Variables.Read(ctx, ws.ID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckTFEVariableSetVariableExists(
	n string, variable *tfe.VariableSetVariable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		vsID := rs.Primary.Attributes["variable_set_id"]
		vs, err := testAccConfiguredClient.Client.VariableSets.Read(ctx, vsID, nil)
		if err != nil {
			return fmt.Errorf(
				"Error retrieving variable set %s: %w", vsID, err)
		}

		v, err := testAccConfiguredClient.Client.VariableSetVariables.Read(ctx, vs.ID, rs.Primary.ID)
		if err != nil {
			return err
		}

		*variable = *v

		return nil
	}
}

func testAccCheckTFEVariableAttributes(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_test" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "value_test" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "some description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryEnv {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != false {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableiSetVariableAttributes(
	variable *tfe.VariableSetVariable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_test" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "value_test" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "some description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryEnv {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != false {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != false {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableAttributesUpdate(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_updated" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "another description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryTerraform {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != true {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != true {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableAttributesUpdate_key_sensitive(
	variable *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != "key_updated_2" {
			return fmt.Errorf("Bad key: %s", variable.Key)
		}

		if variable.Value != "" {
			return fmt.Errorf("Bad value: %s", variable.Value)
		}

		if variable.Description != "another description" {
			return fmt.Errorf("Bad description: %s", variable.Description)
		}

		if variable.Category != tfe.CategoryTerraform {
			return fmt.Errorf("Bad category: %s", variable.Category)
		}

		if variable.HCL != true {
			return fmt.Errorf("Bad HCL: %t", variable.HCL)
		}

		if variable.Sensitive != true {
			return fmt.Errorf("Bad sensitive: %t", variable.Sensitive)
		}

		return nil
	}
}

func testAccCheckTFEVariableIDsNotEqual(
	a, b *tfe.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if a.ID == b.ID {
			return fmt.Errorf("Variables should not have same ID: %s, %s", a.ID, b.ID)
		}

		return nil
	}
}

func testAccCheckTFEVariableDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_variable" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := testAccConfiguredClient.Client.Variables.Read(ctx, rs.Primary.Attributes["workspace_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Variable %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEVariable_basic(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFEVariable_basic_variable_set(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  variable_set_id = tfe_variable_set.foobar.id
}`, rInt)
}

func testAccTFEVariable_update(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_updated"
  value        = "value_updated"
  description  = "another description"
  category     = "terraform"
  hcl          = true
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFEVariable_update_key_sensitive(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_updated_2"
  value        = "value_updated"
  description  = "another description"
  category     = "terraform"
  hcl          = true
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}`, rInt)
}

func testAccTFEVariable_everything(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
  execution_mode = "remote"
}

resource "tfe_variable" "ws_env" {
  key          = "ENV_ONE"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_env_sensitive" {
  key          = "ENV_SENSITIVE"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  sensitive    = true
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform" {
  key          = "key_one"
  value        = "value_test"
  description  = "some description"
  category     = "terraform"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform_hcl" {
  key          = "key_hcl"
  value        = "{ map_key = \"value\" }"
  description  = "some description"
  category     = "terraform"
  hcl          = true
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable" "ws_terraform_no_val" {
  key          = "key_no_val"
  # value absent, defaults to empty string
  description  = "some description"
  category     = "terraform"
  workspace_id = tfe_workspace.foobar.id
}

resource "tfe_variable_set" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "vs_env" {
  key          = "ENV_ONE"
  value        = "value_test"
  description  = "other description"
  category     = "env"
  variable_set_id = tfe_variable_set.foobar.id
}

resource "tfe_variable" "vs_env_sensitive" {
  key          = "ENV_TWO"
  value        = "value_test"
  description  = "other description"
  category     = "env"
  sensitive    = true
  variable_set_id = tfe_variable_set.foobar.id
}

resource "tfe_variable" "vs_terraform" {
  key          = "key_whatever"
  value        = "\"hcl string\""
  description  = "other description"
  category     = "terraform"
  hcl          = true
  variable_set_id = tfe_variable_set.foobar.id
}`, rInt)
}

func testAccTFEVariable_valueWriteOnly(rIntOrg int, rIntVariableValue int, valueVersion int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value_wo        = "%d"
  value_wo_version = %d
  description  = "my description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
  sensitive    = %s
}
`, rIntOrg, rIntVariableValue, valueVersion, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_valueAndValueWO(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value = "%d"
  value_wo        = "%d"
	value_wo_version = 1
  description  = "my description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
  sensitive    = %s
}
`, rIntOrg, rIntVariableValue, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_readablevalue(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "%d"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
  sensitive    = %s
}

resource "tfe_workspace" "downstream-readable" {
  name         = "downstream-readable-${tfe_variable.foobar.readable_value}"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "downstream-nonreadable" {
  name         = "downstream-nonreadable-${tfe_variable.foobar.value}"
  organization = tfe_organization.foobar.id
}
`, rIntOrg, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_varset_readablevalue(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_variable_set" "variable_set" {
  name         = "varset-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "%d"
  description  = "some description"
  category     = "env"
  variable_set_id = tfe_variable_set.variable_set.id
  sensitive    = %s
}

resource "tfe_workspace" "downstream-readable" {
  name         = "downstream-readable-${tfe_variable.foobar.readable_value}"
  organization = tfe_organization.foobar.id
}

resource "tfe_workspace" "downstream-nonreadable" {
  name         = "downstream-nonreadable-${tfe_variable.foobar.value}"
  organization = tfe_organization.foobar.id
}
`, rIntOrg, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_valueWriteOnly_variable_set(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
	name         = "varset-test"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key             = "key_test"
  value_wo        = "%d"
	value_wo_version = 1
  description     = "my description"
  category        = "env"
  variable_set_id = tfe_variable_set.foobar.id
  sensitive       = %s
}
`, rIntOrg, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_valueAndValueWO_variable_set(rIntOrg int, rIntVariableValue int, sensitive bool) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "tst-terraform-%d"
	email = "admin@company.com"
}

resource "tfe_variable_set" "foobar" {
	name         = "varset-test"
	organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key             = "key_test"
  value           = "%d"
  value_wo        = "%d"
	value_wo_version = 1
  description     = "my description"
  category        = "env"
  variable_set_id = tfe_variable_set.foobar.id
  sensitive       = %s
}
`, rIntOrg, rIntVariableValue, rIntVariableValue, strconv.FormatBool(sensitive))
}

func testAccTFEVariable_mutableIdentity_workspace(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable_set" "foobar" {
  name         = "varset-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value        = "value_test"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}
`, rInt)
}

func testAccTFEVariable_mutableIdentity_varset(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable_set" "foobar" {
  name         = "varset-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key             = "key_test"
  value           = "value_test"
  description     = "some description"
  category        = "env"
  variable_set_id = tfe_variable_set.foobar.id
}
`, rInt)
}

func testAccTFEVariable_valueWOOnly(rInt int, rIntVariableValue int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key          = "key_test"
  value_wo     = "%d"
  description  = "some description"
  category     = "env"
  workspace_id = tfe_workspace.foobar.id
}
`, rInt, rIntVariableValue)
}

func testAccTFEVariable_versionOnly(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key               = "key_test"
  value_wo_version  = 1
  description       = "some description"
  category          = "env"
  workspace_id      = tfe_workspace.foobar.id
}
`, rInt)
}

func testAccTFEVariable_valueWithVersion(rInt int) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
  name  = "tst-terraform-%d"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = tfe_organization.foobar.id
}

resource "tfe_variable" "foobar" {
  key               = "key_test"
  value             = "value_test"
  value_wo_version  = 1
  description       = "some description"
  category          = "env"
  workspace_id      = tfe_workspace.foobar.id
}
`, rInt)
}
