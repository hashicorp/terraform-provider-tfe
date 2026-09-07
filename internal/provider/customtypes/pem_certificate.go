// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package customtypes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// PEMCertificateType is a string type for certificates, where two values that
// differ only in PEM armor or line wrapping are equal.
//
// TFE re-armors and re-wraps certs at 64 chars on write, so comparing bytes
// against the configured value spuriously fails. Comparing semantically keeps
// that value in state instead.
type PEMCertificateType struct {
	basetypes.StringType
}

var (
	_ basetypes.StringTypable                    = PEMCertificateType{}
	_ basetypes.StringValuableWithSemanticEquals = PEMCertificateValue{}
)

// String returns a human-readable name for the type.
func (t PEMCertificateType) String() string {
	return "customtypes.PEMCertificateType"
}

// Equal reports whether o is also a PEMCertificateType.
func (t PEMCertificateType) Equal(o attr.Type) bool {
	other, ok := o.(PEMCertificateType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

// ValueType returns the value type this type produces.
func (t PEMCertificateType) ValueType(context.Context) attr.Value {
	return PEMCertificateValue{}
}

// ValueFromString converts a StringValue to a PEMCertificateValue.
func (t PEMCertificateType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return PEMCertificateValue{StringValue: in}, nil
}

// ValueFromTerraform converts a Terraform value to a PEMCertificateValue.
func (t PEMCertificateType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T, expected basetypes.StringValue", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

// PEMCertificateValue is the value for PEMCertificateType.
type PEMCertificateValue struct {
	basetypes.StringValue
}

// NewPEMCertificateValue returns a known value holding v.
func NewPEMCertificateValue(v string) PEMCertificateValue {
	return PEMCertificateValue{StringValue: basetypes.NewStringValue(v)}
}

// NewPEMCertificateNull returns a null value.
func NewPEMCertificateNull() PEMCertificateValue {
	return PEMCertificateValue{StringValue: basetypes.NewStringNull()}
}

// Type returns the type of this value.
func (v PEMCertificateValue) Type(context.Context) attr.Type {
	return PEMCertificateType{}
}

// Equal reports whether o is an identical PEMCertificateValue. Formatting
// differences are handled by StringSemanticEquals, not here.
func (v PEMCertificateValue) Equal(o attr.Value) bool {
	other, ok := o.(PEMCertificateValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals reports whether both values hold the same certificate.
// The framework calls it after a read or a write to decide whether the server
// changed anything beyond formatting.
func (v PEMCertificateValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	other, ok := newValuable.(PEMCertificateValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	return samePEMBody(v.ValueString(), other.ValueString()), diags
}

// Matches any PEM armor line, so legacy labels like X509 CERTIFICATE are
// stripped too.
var pemArmor = regexp.MustCompile(`-{5}(BEGIN|END) [A-Z0-9 ]+-{5}`)

// pemBody strips the armor and all whitespace, leaving just the base64
// payload. Certs get pasted with or without the BEGIN/END markers and wrapped
// at all sorts of widths.
func pemBody(s string) string {
	s = pemArmor.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(s), "")
}

// samePEMBody compares two certificates ignoring armor and whitespace. A chain
// collapses into one run of base64, so its blocks must be in the same order.
func samePEMBody(a, b string) bool {
	return pemBody(a) == pemBody(b)
}
