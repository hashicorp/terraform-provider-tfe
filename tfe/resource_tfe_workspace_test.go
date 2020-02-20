package tfe

import (
	"context"
	"fmt"
	"log"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestPackWorkspaceID(t *testing.T) {
	cases := []struct {
		w   *tfe.Workspace
		id  string
		err bool
	}{
		{
			w: &tfe.Workspace{
				Name: "my-workspace-name",
				Organization: &tfe.Organization{
					Name: "my-org-name",
				},
			},
			id:  "my-org-name/my-workspace-name",
			err: false,
		},
		{
			w: &tfe.Workspace{
				Name: "my-workspace-name",
			},
			id:  "",
			err: true,
		},
	}

	for _, tc := range cases {
		id, err := packWorkspaceID(tc.w)
		if (err != nil) != tc.err {
			t.Fatalf("expected error is %t, got %v", tc.err, err)
		}

		if tc.id != id {
			t.Fatalf("expected ID %q, got %q", tc.id, id)
		}
	}
}

func TestUnpackWorkspaceID(t *testing.T) {
	cases := []struct {
		id   string
		org  string
		name string
		err  bool
	}{
		{
			id:   "my-org-name/my-workspace-name",
			org:  "my-org-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "my-workspace-name|my-org-name",
			org:  "my-org-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "some-invalid-id",
			org:  "",
			name: "",
			err:  true,
		},
	}

	for _, tc := range cases {
		org, name, err := unpackWorkspaceID(tc.id)
		if (err != nil) != tc.err {
			t.Fatalf("expected error is %t, got %v", tc.err, err)
		}

		if tc.org != org {
			t.Fatalf("expected organization %q, got %q", tc.org, org)
		}

		if tc.name != name {
			t.Fatalf("expected name %q, got %q", tc.name, name)
		}
	}
}

func TestAccTFEWorkspace_basic(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "0"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_monorepo(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_monorepo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceMonorepoAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-monorepo"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", "/db"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_renamed(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},

			{
				PreConfig: testAccCheckTFEWorkspaceRename,
				Config:    testAccTFEWorkspace_renamed,
				PlanOnly:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
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
			{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-test"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},

			{
				Config: testAccTFEWorkspace_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdated(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", "terraform/test"),
				),
			},
			{
				Config: testAccTFEWorkspace_updateWorkingDirectory,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesUpdatedWorkingDirectory(workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "name", "workspace-updated"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "auto_apply", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "operations", "true"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "queue_all_runs", "false"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "terraform_version", "0.11.1"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.#", "2"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.0", "/modules"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "trigger_prefixes.1", "/shared"),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "working_directory", ""),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_updateFileTriggers(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "true"),
				),
			},

			{
				Config: testAccTFEWorkspace_basicFileTriggersOff,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					resource.TestCheckResourceAttr(
						"tfe_workspace.foobar", "file_triggers_enabled", "false"),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_sshKey(t *testing.T) {
	workspace := &tfe.Workspace{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
				),
			},

			{
				Config: testAccTFEWorkspace_sshKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributesSSHKey(workspace),
					resource.TestCheckResourceAttrSet(
						"tfe_workspace.foobar", "ssh_key_id"),
				),
			},

			{
				Config: testAccTFEWorkspace_noSSHKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEWorkspaceExists(
						"tfe_workspace.foobar", workspace),
					testAccCheckTFEWorkspaceAttributes(workspace),
				),
			},
		},
	})
}

func TestAccTFEWorkspace_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspace_basic,
			},

			{
				ResourceName:      "tfe_workspace.foobar",
				ImportState:       true,
				ImportStateVerify: true,
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

		// Get the organization and workspace name.
		organization, name, err := unpackWorkspaceID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error unpacking workspace ID: %v", err)
		}

		w, err := tfeClient.Workspaces.Read(ctx, organization, name)
		if err != nil {
			return err
		}

		id, err := packWorkspaceID(w)
		if err != nil {
			return fmt.Errorf("Error creating ID for workspace %s: %v", name, err)
		}

		if id != rs.Primary.ID {
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

		if workspace.AutoApply != true {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.Operations != true {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
		}

		if workspace.QueueAllRuns != true {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
		}

		if workspace.SSHKey != nil {
			return fmt.Errorf("Bad SSH key: %v", workspace.SSHKey)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceMonorepoAttributes(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-monorepo" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.FileTriggersEnabled != true {
			return fmt.Errorf("Bad file triggers enabled: %t", workspace.FileTriggersEnabled)
		}

		triggerPrefixes := []string{"/modules", "/shared"}
		if len(workspace.TriggerPrefixes) != len(triggerPrefixes) {
			return fmt.Errorf("Bad trigger prefixes length: %d", len(workspace.TriggerPrefixes))
		}
		for i := range triggerPrefixes {
			if workspace.TriggerPrefixes[i] != triggerPrefixes[i] {
				return fmt.Errorf("Bad trigger prefixes %v", workspace.TriggerPrefixes)
			}
		}

		if workspace.WorkingDirectory != "/db" {
			return fmt.Errorf("Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceRename() {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	w, err := tfeClient.Workspaces.Update(
		context.Background(),
		"tst-terraform",
		"workspace-test",
		tfe.WorkspaceUpdateOptions{Name: tfe.String("renamed-out-of-band")},
	)
	if err != nil {
		log.Fatalf("Could not rename the workspace out of band: %v", err)
	}

	if w.Name != "renamed-out-of-band" {
		log.Fatalf("Failed to rename the workspace out of band: %v", err)
	}
}

func testAccCheckTFEWorkspaceAttributesUpdated(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != false {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.Operations != true {
			return fmt.Errorf("Bad operations: %t", workspace.Operations)
		}

		if workspace.QueueAllRuns != false {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
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

func testAccCheckTFEWorkspaceAttributesUpdatedWorkingDirectory(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.Name != "workspace-updated" {
			return fmt.Errorf("Bad name: %s", workspace.Name)
		}

		if workspace.AutoApply != false {
			return fmt.Errorf("Bad auto apply: %t", workspace.AutoApply)
		}

		if workspace.Operations != true {
			return fmt.Errorf("Here Bad operations: %t", workspace.Operations)
		}

		if workspace.QueueAllRuns != false {
			return fmt.Errorf("Bad queue all runs: %t", workspace.QueueAllRuns)
		}

		if workspace.TerraformVersion != "0.11.1" {
			return fmt.Errorf("Bad Terraform version: %s", workspace.TerraformVersion)
		}

		if workspace.WorkingDirectory != "" {
			return fmt.Errorf("Today Bad working directory: %s", workspace.WorkingDirectory)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceAttributesSSHKey(
	workspace *tfe.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if workspace.SSHKey == nil {
			return fmt.Errorf("Bad SSH key: %v", workspace.SSHKey)
		}

		return nil
	}
}

func testAccCheckTFEWorkspaceDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_workspace" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		// Get the organization and workspace name.
		organization, name, err := unpackWorkspaceID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error unpacking workspace ID: %v", err)
		}

		_, err = tfeClient.Workspaces.Read(ctx, organization, name)
		if err == nil {
			return fmt.Errorf("Workspace %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccTFEWorkspace_basic = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`

const testAccTFEWorkspace_basicFileTriggersOff = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-test"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = true
  file_triggers_enabled = false
}`

const testAccTFEWorkspace_monorepo = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-monorepo"
  organization          = "${tfe_organization.foobar.id}"
  file_triggers_enabled = true
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "/db"
}`

const testAccTFEWorkspace_renamed = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name         = "renamed-out-of-band"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`

const testAccTFEWorkspace_update = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = false
  file_triggers_enabled = true
  queue_all_runs        = false
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
  working_directory     = "terraform/test"
}`

const testAccTFEWorkspace_updateWorkingDirectory = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_workspace" "foobar" {
  name                  = "workspace-updated"
  organization          = "${tfe_organization.foobar.id}"
  auto_apply            = false
  file_triggers_enabled = true
  queue_all_runs        = false
  terraform_version     = "0.11.1"
  trigger_prefixes      = ["/modules", "/shared"]
}`

const testAccTFEWorkspace_sshKey = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = "${tfe_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
  ssh_key_id   = "${tfe_ssh_key.foobar.id}"
}`

const testAccTFEWorkspace_noSSHKey = `
resource "tfe_organization" "foobar" {
  name  = "tst-terraform"
  email = "admin@company.com"
}

resource "tfe_ssh_key" "foobar" {
  name         = "ssh-key-test"
  organization = "${tfe_organization.foobar.id}"
  key          = "SSH-KEY-CONTENT"
}

resource "tfe_workspace" "foobar" {
  name         = "workspace-test"
  organization = "${tfe_organization.foobar.id}"
  auto_apply   = true
}`
