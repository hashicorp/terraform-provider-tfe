package tfe

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceTFEVariable struct{}

// Metadata implements resource.Resource
func (*resourceTFEVariable) Metadata(_ context.Context, _ resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = "tfe_variable"
}

// Schema implements resource.Resource
func (*resourceTFEVariable) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Attributes:          map[string]schema.Attribute{},
		Blocks:              map[string]schema.Block{},
		Description:         "",
		MarkdownDescription: "",
		DeprecationMessage:  "",
		Version:             0,
	}
}

// Create implements resource.Resource
func (*resourceTFEVariable) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

// Delete implements resource.Resource
func (*resourceTFEVariable) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}

// Read implements resource.Resource
func (*resourceTFEVariable) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
	panic("unimplemented")
}

// Update implements resource.Resource
func (*resourceTFEVariable) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}

// Compile-time interface check
var _ resource.Resource = &resourceTFEVariable{}

// NewResourceVariable is a resource function for the framework provider.
func NewResourceVariable() resource.Resource {
	return &resourceTFEVariable{}
}
