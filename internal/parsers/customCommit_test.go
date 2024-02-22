package parsers_test

import (
	"reflect"
	"testing"

	"github.com/happyagosmith/jig/internal/parsers"
)

func TestCustomParser_Parse(t *testing.T) {
	parser := parsers.NewCustom(parsers.WithPattern(`\[(?P<scope>[^\]]*)\](?P<subject>.*)`))

	tests := []struct {
		name   string
		commit string
		want   *parsers.ConventionalCommit
	}{
		{
			name:   "Test with valid commit",
			commit: "[AAA-123] add new feature",
			want:   &parsers.ConventionalCommit{Scope: "AAA-123", Subject: "add new feature"},
		},
		{
			name:   "Test with invalid commit",
			commit: "invalid commit message",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Parse(tt.commit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CustomParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
