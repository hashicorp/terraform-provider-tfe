// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

// NOTE: This is a legacy resource and should be migrated to the Plugin
// Framework if substantial modifications are planned. See
// docs/new-resources.md if planning to use this code as boilerplate for
// a new resource.

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-tfe"
	tfeV2 "github.com/hashicorp/go-tfe/v2"
	"github.com/hashicorp/go-tfe/v2/api/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-tfe/internal/provider/helpers"
)

func resourceTFEPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Manages policies." +
			"\n\nPolicies are rules enforced on Terraform runs. You can use policies to validate that the Terraform plan complies with security rules and best practices. Two policy-as-code frameworks are integrated with Terraform Enterprise: Sentinel and Open Policy Agent (OPA)." +
			"\n\nPolicies are configured on a per-organization level and are organized and grouped into policy sets, which define the workspaces on which policies are enforced during runs.",

		Create: resourceTFEPolicyCreate,
		Read:   resourceTFEPolicyRead,
		Update: resourceTFEPolicyUpdate,
		Delete: resourceTFEPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEPolicyImporter,
		},

		CustomizeDiff: customizeDiffIfProviderDefaultOrganizationChanged,

		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"hostname": {
						Type:              schema.TypeString,
						OptionalForImport: true,
					},
					"organization": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
				}
			},
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the policy.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"name": {
				Description: "The name of the policy.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"description": {
				Description: "Text describing the policy's purpose.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"organization": {
				Description: "Name of the organization that this policy belongs to. If omitted, organization must be defined in the provider config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"kind": {
				Description: "The policy-as-code framework for the policy. Valid values are `sentinel` and `opa`. Defaults to `sentinel` if not provided.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     string(tfe.Sentinel),
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.OPA),
						string(tfe.Sentinel),
					}, false),
			},

			"query": {
				Description: "The OPA query to identify a specific policy rule that needs to run within your Rego code. Required for OPA policies.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"policy": {
				Description: "Text of a valid Sentinel or OPA policy.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"enforce_mode": {
				Type: schema.TypeString,
				Description: fmt.Sprintf(
					"The enforcement configuration of the policy. For Sentinel, valid values are %s; defaults to `%s`. For OPA, valid values are %s; defaults to `%s`.", sentenceList(
						sentinelPolicyEnforcementLevels(),
						"`",
						"`",
						"and"),
					string(tfe.EnforcementSoft),
					sentenceList(
						opaPolicyEnforcementLevels(),
						"`",
						"`",
						"and"),
					string(tfe.EnforcementAdvisory)),
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.EnforcementAdvisory),
						string(tfe.EnforcementHard),
						string(tfe.EnforcementSoft),
						string(tfe.EnforcementMandatory),
					},
					false,
				),
			},
		},
	}
}

func sentinelPolicyEnforcementLevels() []string {
	return []string{
		string(tfe.EnforcementHard),
		string(tfe.EnforcementSoft),
		string(tfe.EnforcementAdvisory),
	}
}

func opaPolicyEnforcementLevels() []string {
	return []string{
		string(tfe.EnforcementMandatory),
		string(tfe.EnforcementAdvisory),
	}
}

func resourceTFEPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	// Get the name and organization.
	name := d.Get("name").(string)
	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	var kind string
	if vKind, ok := d.GetOk("kind"); ok {
		kind = vKind.(string)
	}

	attrs := models.NewPolicies_attributes()
	attrs.SetName(ptr(name))
	parsedKind, _ := models.ParsePolicies_attributes_kind(kind)
	if parsedKind != nil {
		k := parsedKind.(*models.Policies_attributes_kind)
		attrs.SetKind(k)
	}

	if desc, ok := d.GetOk("description"); ok {
		attrs.SetDescription(ptr(desc.(string)))
	}

	var createErr error
	switch tfe.PolicyKind(kind) {
	case tfe.Sentinel:
		attrs = createSentinelPolicyAttrs(attrs, d)
	case tfe.OPA:
		attrs, createErr = createOPAPolicyAttrs(attrs, d)
	default:
		createErr = fmt.Errorf(
			"unsupported policy kind %s: has to be one of [%s, %s]", kind, string(tfe.Sentinel), string(tfe.OPA))
	}
	if createErr != nil {
		return createErr
	}

	body := models.NewPolicies()
	body.SetAttributes(attrs)
	env := models.NewPoliciesEnvelope()
	env.SetData(body)

	log.Printf("[DEBUG] Create %s policy %s for organization: %s", kind, name, organization)
	policyEnv, err := config.ClientV2.API.Organizations().ByOrganization_name(organization).Policies().Post(ctx, env, nil)
	if err != nil {
		return fmt.Errorf(
			"Error creating %s policy %s for organization %s: %w", kind, name, organization, err)
	}

	policyID := valueOrZero(policyEnv.GetData().GetId())
	d.SetId(policyID)

	err = helpers.WriteTFEIdentityWithOrg(d, policyID, organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Upload %s policy %s for organization: %s", kind, name, organization)
	_, err = config.ClientV2.API.Policies().ByPolicy_id(policyID).Upload().Put(ctx, []byte(d.Get("policy").(string)), nil)
	if err != nil {
		return fmt.Errorf(
			"Error uploading %s policy %s for organization %s: %w", kind, name, organization, err)
	}

	return resourceTFEPolicyRead(d, meta)
}

func createOPAPolicyAttrs(attrs *models.Policies_attributes, d *schema.ResourceData) (*models.Policies_attributes, error) {
	name := d.Get("name").(string)
	path := name + ".rego"
	enforceEntry := models.NewPolicies_attributes_enforce()
	enforceEntry.SetPath(ptr(path))
	if v, ok := d.GetOk("enforce_mode"); !ok {
		enforceEntry.SetMode(ptr(string(getDefaultEnforcementMode(tfe.OPA))))
	} else {
		enforceEntry.SetMode(ptr(v.(string)))
	}
	//nolint:staticcheck // this is still used by TFE versions older than 202306-1
	attrs.SetEnforce([]models.Policies_attributes_enforceable{enforceEntry})

	vQuery, ok := d.GetOk("query")
	if !ok {
		return attrs, fmt.Errorf("missing query for OPA policy")
	}
	attrs.SetQuery(ptr(vQuery.(string)))
	return attrs, nil
}

func createSentinelPolicyAttrs(attrs *models.Policies_attributes, d *schema.ResourceData) *models.Policies_attributes {
	name := d.Get("name").(string)
	path := name + ".sentinel"
	enforceEntry := models.NewPolicies_attributes_enforce()
	enforceEntry.SetPath(ptr(path))
	if v, ok := d.GetOk("enforce_mode"); !ok {
		enforceEntry.SetMode(ptr(string(getDefaultEnforcementMode(tfe.Sentinel))))
	} else {
		enforceEntry.SetMode(ptr(v.(string)))
	}
	//nolint:staticcheck // this is still used by TFE versions older than 202306-1
	attrs.SetEnforce([]models.Policies_attributes_enforceable{enforceEntry})
	return attrs
}

func getDefaultEnforcementMode(kind tfe.PolicyKind) tfe.EnforcementLevel {
	switch kind {
	case tfe.Sentinel:
		return tfe.EnforcementSoft

	case tfe.OPA:
		return tfe.EnforcementAdvisory

	default:
		return ""
	}
}

func resourceTFEPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Read policy: %s", d.Id())
	policyEnv, err := config.ClientV2.API.Policies().ByPolicy_id(d.Id()).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			log.Printf("[DEBUG] Policy %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Policy %s: %w", d.Id(), err)
	}

	policy := policyEnv.GetData()
	attrs := policy.GetAttributes()

	d.Set("name", valueOrZero(attrs.GetName()))
	d.Set("description", valueOrZero(attrs.GetDescription()))
	d.Set("kind", enumStringOrEmpty(attrs.GetKind()))

	//nolint:staticcheck // this is still used by TFE versions older than 202306-1
	for _, e := range attrs.GetEnforce() {
		d.Set("enforce_mode", valueOrZero(e.GetMode()))
		break
	}

	if q := attrs.GetQuery(); q != nil {
		d.Set("query", *q)
	}

	content, err := config.ClientV2.API.Policies().ByPolicy_id(d.Id()).Download().Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error downloading policy %s: %w", d.Id(), err)
	}
	d.Set("policy", string(content))

	organization, err := config.schemaOrDefaultOrganization(d)
	if err != nil {
		return err
	}

	err = helpers.WriteTFEIdentityWithOrg(d, d.Id(), organization, config.Client.BaseURL().Host)
	if err != nil {
		return err
	}

	return nil
}

func resourceTFEPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	var kind string
	if v, ok := d.GetOk("kind"); ok {
		kind = v.(string)
	}

	// nolint:nestif
	if d.HasChange("description") || d.HasChange("enforce_mode") || d.HasChange("query") {
		attrs := models.NewPolicies_attributes()

		if desc, ok := d.GetOk("description"); ok {
			attrs.SetDescription(ptr(desc.(string)))
		}

		if d.HasChange("enforce_mode") {
			path := d.Get("name").(string) + ".sentinel"
			if kind == string(tfe.OPA) {
				path = d.Get("name").(string) + ".rego"
			}
			enforceEntry := models.NewPolicies_attributes_enforce()
			enforceEntry.SetPath(ptr(path))
			enforceEntry.SetMode(ptr(d.Get("enforce_mode").(string)))
			//nolint:staticcheck // this is still used by TFE versions older than 202306-1
			attrs.SetEnforce([]models.Policies_attributes_enforceable{enforceEntry})
		}

		if query, ok := d.GetOk("query"); ok {
			attrs.SetQuery(ptr(query.(string)))
		}

		body := models.NewPolicies()
		body.SetAttributes(attrs)
		env := models.NewPoliciesEnvelope()
		env.SetData(body)

		log.Printf("[DEBUG] Update configuration for %s policy: %s", kind, d.Id())
		_, err := config.ClientV2.API.Policies().ByPolicy_id(d.Id()).Patch(ctx, env, nil)
		if err != nil {
			return fmt.Errorf(
				"Error updating configuration for %s policy %s: %w", kind, d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		vKind := d.Get("kind").(string)
		log.Printf("[DEBUG] Update %s policy: %s", vKind, d.Id())
		_, err := config.ClientV2.API.Policies().ByPolicy_id(d.Id()).Upload().Put(ctx, []byte(d.Get("policy").(string)), nil)
		if err != nil {
			return fmt.Errorf("Error updating %s policy %s: %w", vKind, d.Id(), err)
		}
	}

	return resourceTFEPolicyRead(d, meta)
}

func resourceTFEPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	log.Printf("[DEBUG] Delete policy: %s", d.Id())
	err := config.ClientV2.API.Policies().ByPolicy_id(d.Id()).Delete(ctx, nil)
	if err != nil {
		if errors.Is(err, tfeV2.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("Error deleting policy %s: %w", d.Id(), err)
	}

	return nil
}

func resourceTFEPolicyImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// First we'll check for an identity
	identity, err := d.Identity()
	if err != nil {
		return nil, fmt.Errorf("error reading policy identity: %w", err)
	}

	if externalID := identity.Get("id").(string); externalID != "" {
		// We are importing by identity
		// This only supported when using an import block, since import blocks
		// are the only way to specify an identity. Importing via TF CLI does
		// not support specifying an identity.
		d.SetId(externalID)
		orgName := identity.Get("organization").(string)
		err = d.Set("organization", orgName)
		if err != nil {
			return nil, fmt.Errorf("could not set organization name %s on policy: %w", orgName, err)
		}

		// Exit early
		return []*schema.ResourceData{d}, nil
	}

	// Otherwise we are using legacy import prefix
	s := strings.SplitN(d.Id(), "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid policy import format: %s (expected <ORGANIZATION>/<POLICY ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	d.Set("organization", s[0])
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
