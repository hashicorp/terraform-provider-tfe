package tfe

import (
	"context"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
)

func TestFetchVariableSetExternalID(t *testing.T) {
	tests := map[string]struct {
		def  string
		want string
		err  bool
	}{
		"non exisiting organization": {
			"not-an-org/variable_set",
			"",
			true,
		},
		"non exisiting variable_set": {
			"hashicorp/not-a-variable_set",
			"",
			true,
		},
		"found variable_set": {
			"hashicorp/a-variable_set",
			"vs-123",
			false,
		},
	}

	client := testTfeClient(t, testClientOptions{defaultVariableSetID: "vs-123"})
	name := "a-variable_set"
	client.VariableSets.Create(nil, "hashicorp", tfe.VariableSetCreateOptions{
		Name: &name,
	})

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchVariableSetExternalID(test.def, client)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if got != test.want {
				t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
