package model_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/model"
)

func TestDeprecation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "not deprecated",
			doc:  "Circle is a round shape.\n",
			want: "",
		},
		{
			name: "own paragraph",
			doc:  "Legacy is the old version.\n\nDeprecated: Use Version instead.\n",
			want: "Use Version instead.",
		},
		{
			name: "multi-line message",
			doc:  "Legacy is the old version.\n\nDeprecated: Use Version instead. Legacy is kept\nonly for old callers.\n",
			want: "Use Version instead. Legacy is kept only for old callers.",
		},
		{
			name: "first paragraph",
			doc:  "Deprecated: Gone.\n\nMore text.\n",
			want: "Gone.",
		},
		{
			name: "prefix inside a paragraph",
			doc:  "Legacy is old. Deprecated: not really.\n",
			want: "",
		},
		{
			name: "empty",
			doc:  "",
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			msg := model.Deprecation(tc.doc)

			// Assert
			if got, want := msg, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Deprecation(...) = %q, want %q", got, want)
			}
		})
	}
}
