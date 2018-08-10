package tfe

import (
	"fmt"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTFEWorkspace_basic(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspoce.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_update(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspoce.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "working_directory", ""),
				),
			},

			resource.TestStep{
				Config: testAccTFEWorkspace_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspoce.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"tfe_workspoce.foobar", "working_directory", "terraform/test"),
				),
			},
		},
	})
}

func testAccCheckTFEWorkspaceExists(
	n string, workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the name and organization.
		name, organization := unpackWorkspaceID(rs.Primary.ID)

		w, err := tfeClient.Workspaces.Read(ctx, organization, name)
		if err != nil {
			return err
		}

		if packWorkspaceID(w) != rs.Primary.ID {
			return fmt.Errorf("Workspace not found")
		}

		*workspace = *w

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-test" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != false {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributesUpdated(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != true {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.TerraformVersion != "0.11.1" {
			return fmt.Errorf("Bad Terraform version: %s", workspace.TerraformVersion)
		}

		if workspace.WorkingDirectory != "terraform/test" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_workspoce" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the name and organization.
		name, organization := unpackWorkspaceID(rs.Primary.ID)

		_, err := tfeClient.Workspaces.Read(ctx, organization, name)
		if err == nil {
			return fmt.Errorf("Workspace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEWorkspace_basic = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
	email = "admin@company.com"
}

resource "tfe_workspoce" "foobar" {
	name = "workspace-test"
	organization = "${tfe_organization.foobar.id}"
}`

const testAccTFEWorkspace_update = `
resource "tfe_organization" "foobar" {
  name = "terraform-test"
	email = "admin@company.com"
}

resource "tfe_workspoce" "foobar" {
	name = "workspace-updated"
	organization = "${tfe_organization.foobar.id}"
	auto_apply = true
	terraform_version = "0.11.1"
	working_directory = "terraform/test"
}`
