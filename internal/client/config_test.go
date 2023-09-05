// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"os"
	"testing"
)

func TestConfig_locateConfigFile(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalTfCliConfigFile := os.Getenv("TF_CLI_CONFIG_FILE")
	originalTerraformConfig := os.Getenv("TERRAFORM_CONFIG")
	reset := func() {
		os.Setenv("HOME", originalHome)
		if originalTfCliConfigFile != "" {
			os.Setenv("TF_CLI_CONFIG_FILE", originalTfCliConfigFile)
		} else {
			os.Unsetenv("TF_CLI_CONFIG_FILE")
		}
		if originalTerraformConfig != "" {
			os.Setenv("TERRAFORM_CONFIG", originalTerraformConfig)
		} else {
			os.Unsetenv("TERRAFORM_CONFIG")
		}
	}
	defer reset()

	// Use a predictable value for $HOME
	os.Setenv("HOME", "/Users/someone")

	setup := func(tfCliConfigFile, terraformConfig string) {
		os.Setenv("TF_CLI_CONFIG_FILE", tfCliConfigFile)
		os.Setenv("TERRAFORM_CONFIG", terraformConfig)
	}

	cases := map[string]struct {
		tfCliConfigFile string
		terraformConfig string
		result          string
	}{
		"has TF_CLI_CONFIG_FILE": {
			tfCliConfigFile: "~/.terraform_alternate/terraformrc",
			terraformConfig: "",
			result:          "~/.terraform_alternate/terraformrc",
		},
		"has TERRAFORM_CONFIG": {
			tfCliConfigFile: "",
			terraformConfig: "~/.terraform_alternate_rc",
			result:          "~/.terraform_alternate_rc",
		},
		"has both env vars": {
			tfCliConfigFile: "~/.from_TF_CLI",
			terraformConfig: "~/.from_TERRAFORM_CONFIG",
			result:          "~/.from_TF_CLI",
		},
		"has neither env var": {
			tfCliConfigFile: "",
			terraformConfig: "",
			result:          "/Users/someone/.terraformrc", // expect tests run on unix
		},
	}

	for name, tc := range cases {
		setup(tc.tfCliConfigFile, tc.terraformConfig)

		fileResult := locateConfigFile()
		if tc.result != fileResult {
			t.Fatalf("%s: expected config file at %s, got %s", name, tc.result, fileResult)
		}
	}
}

func TestConfig_cliConfig(t *testing.T) {
	// This only tests the behavior of merging various combinations of
	// (credentials file, .terraformrc file, absent). Locating the .terraformrc
	// file is tested separately.
	originalHome := os.Getenv("HOME")
	originalTfCliConfigFile := os.Getenv("TF_CLI_CONFIG_FILE")
	reset := func() {
		os.Setenv("HOME", originalHome)
		if originalTfCliConfigFile != "" {
			os.Setenv("TF_CLI_CONFIG_FILE", originalTfCliConfigFile)
		} else {
			os.Unsetenv("TF_CLI_CONFIG_FILE")
		}
	}
	defer reset()

	// Summary of fixtures: the credentials file and terraformrc file each have
	// credentials for two hosts, they both have credentials for app.terraform.io,
	// and the terraformrc also has one service discovery override.
	hasCredentials := "test-fixtures/cli-config-files/home"
	noCredentials := "test-fixtures/cli-config-files/no-credentials"
	terraformrc := "test-fixtures/cli-config-files/terraformrc"
	noTerraformrc := "test-fixtures/cli-config-files/no-terraformrc"
	tokenFromTerraformrc := "something.atlasv1.prod_rc_file"
	tokenFromCredentials := "something.atlasv1.prod_credentials_file"

	cases := map[string]struct {
		home             string
		rcfile           string
		expectCount      int
		expectProdToken  string
		expectHostsCount int
	}{
		"both main config and credentials JSON": {
			home:             hasCredentials,
			rcfile:           terraformrc,
			expectCount:      3,
			expectProdToken:  tokenFromTerraformrc,
			expectHostsCount: 1,
		},
		"only main config": {
			home:             noCredentials,
			rcfile:           terraformrc,
			expectCount:      2,
			expectProdToken:  tokenFromTerraformrc,
			expectHostsCount: 1,
		},
		"only credentials JSON": {
			home:             hasCredentials,
			rcfile:           noTerraformrc,
			expectCount:      2,
			expectProdToken:  tokenFromCredentials,
			expectHostsCount: 0,
		},
		"neither file": {
			home:             noCredentials,
			rcfile:           noTerraformrc,
			expectCount:      0,
			expectProdToken:  "",
			expectHostsCount: 0,
		},
	}

	for name, tc := range cases {
		os.Setenv("HOME", tc.home)
		os.Setenv("TF_CLI_CONFIG_FILE", tc.rcfile)
		config := cliConfig()
		credentialsCount := len(config.Credentials)
		if credentialsCount != tc.expectCount {
			t.Fatalf("%s: expected %d credentials, got %d", name, tc.expectCount, credentialsCount)
		}
		prodToken := ""
		if config.Credentials["app.terraform.io"] != nil {
			prodToken = config.Credentials["app.terraform.io"]["token"].(string)
		}
		if prodToken != tc.expectProdToken {
			t.Fatalf("%s: expected %s as prod token, got %s", name, tc.expectProdToken, prodToken)
		}
		hostsCount := len(config.Hosts)
		if hostsCount != tc.expectHostsCount {
			t.Fatalf("%s: expected %d `host` blocks in the final config, got %d", name, tc.expectHostsCount, hostsCount)
		}
	}
}
