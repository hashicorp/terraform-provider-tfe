func resourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
		Description: "An agent pool represents a group of agents..." + 
			"\n\n ~> **NOTE:** This resource requires  ... ",

		Create: resourceTFEAgentPoolCreate,
		Read:   resourceTFEAgentPoolRead,
		Update: resourceTFEAgentPoolUpdate,
		Delete: resourceTFEAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
				Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}