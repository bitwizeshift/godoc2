package search_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/search"
)

func TestIndex_Write(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		entries []search.Entry
		want    string
	}{
		{
			name:    "empty",
			entries: nil,
			want:    "window.godoc2SearchIndex = [];\n",
		},
		{
			name: "entries",
			entries: []search.Entry{
				{Name: "Circle", Kind: "type", Package: "example.com/sample", Path: "example.com/sample/Circle.html"},
				{Name: "Circle.Area", Kind: "method", Package: "example.com/sample", Path: "example.com/sample/Circle.Area.html"},
			},
			want: `window.godoc2SearchIndex = [` +
				`{"name":"Circle","kind":"type","package":"example.com/sample","path":"example.com/sample/Circle.html"},` +
				`{"name":"Circle.Area","kind":"method","package":"example.com/sample","path":"example.com/sample/Circle.Area.html"}` +
				"];\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &search.Index{}
			for _, e := range tc.entries {
				sut.Add(e)
			}
			var out strings.Builder

			// Act
			err := sut.Write(&out)
			text := out.String()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Index.Write(...) = %v, want nil", got)
			}
			if got, want := text, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Index.Write(...) = %q, want %q", got, want)
			}
		})
	}
}
