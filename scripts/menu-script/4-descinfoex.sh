func resourceTFEAgentPool() *schema.Resource {
	return &schema.Resource{
[1]		Description: "An agent pool represents a group of agents..." + 
[2]			"\n\n ~> This resource requires  ... ",

		Create: resourceTFEAgentPoolCreate,
		Read:   resourceTFEAgentPoolRead,
		Update: resourceTFEAgentPoolUpdate,
		Delete: resourceTFEAgentPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
[1]			Description: "Name of the agent pool.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"organization": {
[1]			Description: "Name of the organization.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}