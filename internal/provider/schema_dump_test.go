// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestSchemaDump writes the protocol-level schema of selected resources and
// data sources as sorted "path = value" lines. It is a tool for
// scripts/schema-diff.sh, which runs it before and after a plugin framework
// migration and compares the output (docs/framework-migration.md).
//
// It is skipped unless SCHEMA_DUMP is set:
//
//	SCHEMA_DUMP=tfe_team,data.tfe_team SCHEMA_DUMP_OUT=/tmp/schema.txt \
//	  go test ./internal/provider -run '^TestSchemaDump$' -count=1
//
// Names are resource types; prefix with "data." for data sources. Output
// covers what Terraform sees: types, required/optional/computed, sensitive,
// deprecated, write-only, descriptions, block nesting and item limits, schema
// version, and the resource identity schema. Provider-internal behavior
// (plan modifiers, defaults, validators) is not part of the protocol schema
// and is not included.
func TestSchemaDump(t *testing.T) {
	names := os.Getenv("SCHEMA_DUMP")
	if names == "" {
		t.Skip("set SCHEMA_DUMP to dump schemas (see scripts/schema-diff.sh)")
	}
	out := os.Getenv("SCHEMA_DUMP_OUT")
	if out == "" {
		t.Fatal("SCHEMA_DUMP_OUT must be set to the output file path")
	}

	ctx := context.Background()
	server, err := testAccMuxedProviders["tfe"]()
	if err != nil {
		t.Fatalf("creating provider server: %v", err)
	}
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("getting provider schema: %v", err)
	}
	for _, d := range schemas.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("provider schema error: %s: %s", d.Summary, d.Detail)
		}
	}
	identities, err := server.GetResourceIdentitySchemas(ctx, &tfprotov6.GetResourceIdentitySchemasRequest{})
	if err != nil {
		t.Fatalf("getting resource identity schemas: %v", err)
	}

	var lines []string
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		dumped, err := dumpSchemaLines(name, schemas, identities)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, dumped...)
	}
	sort.Strings(lines)

	if err := os.WriteFile(out, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("writing %s: %v", out, err)
	}
}

func dumpSchemaLines(name string, schemas *tfprotov6.GetProviderSchemaResponse, identities *tfprotov6.GetResourceIdentitySchemasResponse) ([]string, error) {
	var lines []string
	add := func(path string, value any) {
		if s, ok := value.(string); ok {
			value = strconv.Quote(strings.TrimSpace(s))
		}
		lines = append(lines, fmt.Sprintf("%s = %v", path, value))
	}

	if typeName, ok := strings.CutPrefix(name, "data."); ok {
		s, found := schemas.DataSourceSchemas[typeName]
		if !found {
			return nil, fmt.Errorf("data source %q not found in provider schema", typeName)
		}
		prefix := "data." + typeName
		add(prefix+".version", s.Version)
		dumpBlock(add, prefix, s.Block)
		return lines, nil
	}

	s, found := schemas.ResourceSchemas[name]
	if !found {
		return nil, fmt.Errorf("resource %q not found in provider schema", name)
	}
	prefix := "resource." + name
	add(prefix+".version", s.Version)
	dumpBlock(add, prefix, s.Block)

	if id, ok := identities.IdentitySchemas[name]; ok && id != nil {
		add(prefix+".identity.version", id.Version)
		for _, a := range id.IdentityAttributes {
			p := prefix + ".identity.attribute." + a.Name
			add(p+".type", typeString(a.Type))
			add(p+".required_for_import", a.RequiredForImport)
			add(p+".optional_for_import", a.OptionalForImport)
			add(p+".description", a.Description)
		}
	}
	return lines, nil
}

func dumpBlock(add func(string, any), prefix string, b *tfprotov6.SchemaBlock) {
	if b == nil {
		return
	}
	add(prefix+".description", b.Description)
	add(prefix+".deprecated", b.Deprecated)
	for _, a := range b.Attributes {
		dumpAttribute(add, prefix+".attribute."+a.Name, a)
	}
	for _, nb := range b.BlockTypes {
		p := prefix + ".block." + nb.TypeName
		add(p+".nesting", nb.Nesting.String())
		add(p+".min_items", nb.MinItems)
		add(p+".max_items", nb.MaxItems)
		dumpBlock(add, p, nb.Block)
	}
}

func dumpAttribute(add func(string, any), p string, a *tfprotov6.SchemaAttribute) {
	add(p+".type", typeString(a.Type))
	add(p+".required", a.Required)
	add(p+".optional", a.Optional)
	add(p+".computed", a.Computed)
	add(p+".sensitive", a.Sensitive)
	add(p+".deprecated", a.Deprecated)
	add(p+".write_only", a.WriteOnly)
	add(p+".description", a.Description)
	if a.NestedType != nil {
		add(p+".nested.nesting", a.NestedType.Nesting.String())
		for _, na := range a.NestedType.Attributes {
			dumpAttribute(add, p+".nested.attribute."+na.Name, na)
		}
	}
}

func typeString(t interface{ String() string }) string {
	if t == nil {
		return ""
	}
	return t.String()
}
