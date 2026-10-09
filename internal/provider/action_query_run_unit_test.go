// Copyright IBM Corp. 2018, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	tfemocks "github.com/hashicorp/go-tfe/mocks"
	"go.uber.org/mock/gomock"
)

func TestQueryRunLogTail(t *testing.T) {
	var long strings.Builder
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&long, "line %d\n", i)
	}

	cases := map[string]struct {
		log  string
		err  error
		want string
	}{
		"short log returned whole": {log: "a\nb\n", want: "a\nb"},
		"long log truncated to tail": {
			log:  long.String(),
			want: strings.TrimSpace(strings.Join(strings.Split(long.String(), "\n")[10:30], "\n")),
		},
		"log error returns empty": {err: errors.New("nope"), want: ""},
		"empty log returns empty": {log: "", want: ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mock := tfemocks.NewMockQueryRuns(ctrl)
			var r *strings.Reader
			if tc.err == nil {
				r = strings.NewReader(tc.log)
			}
			if tc.err != nil {
				mock.EXPECT().Logs(gomock.Any(), "qr-123").Return(nil, tc.err)
			} else {
				mock.EXPECT().Logs(gomock.Any(), "qr-123").Return(r, nil)
			}

			got := queryRunLogTail(context.Background(), &tfe.Client{QueryRuns: mock}, "qr-123")
			if got != tc.want {
				t.Fatalf("queryRunLogTail() = %q, want %q", got, tc.want)
			}
		})
	}
}
