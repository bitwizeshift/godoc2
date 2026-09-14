package props_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/props"
)

func TestBadges(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []props.Badge
	}{
		{
			name: "struct",
			typ:  loadertest.Type(t, "", "Circle"),
			want: []props.Badge{
				{Label: "size", Value: "24 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
			},
		},
		{
			name: "sealed interface",
			typ:  loadertest.Type(t, "", "Shape"),
			want: []props.Badge{
				{Label: "size", Value: "16 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
				{Label: "sealed"},
			},
		},
		{
			name: "open interface",
			typ:  loadertest.Type(t, "", "Named"),
			want: []props.Badge{
				{Label: "size", Value: "16 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
			},
		},
		{
			name: "large array struct",
			typ:  loadertest.Type(t, "", "Big"),
			want: []props.Badge{
				{Label: "size", Value: "128 bytes"},
				{Label: "align", Value: "1"},
				{Label: "large"},
				{Label: "comparable"},
			},
		},
		{
			name: "struct with mutex field",
			typ:  loadertest.Type(t, "", "Counter"),
			want: []props.Badge{
				{Label: "size", Value: "16 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
				{Label: "noncopiable"},
			},
		},
		{
			name: "array of noncopiable structs",
			typ:  loadertest.Type(t, "", "Grid"),
			want: []props.Badge{
				{Label: "size", Value: "64 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
				{Label: "noncopiable"},
			},
		},
		{
			name: "generic struct",
			typ:  loadertest.Type(t, "", "Stack"),
			want: nil,
		},
		{
			name: "alias",
			typ:  loadertest.Type(t, "", "ID"),
			want: []props.Badge{
				{Label: "size", Value: "16 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
			},
		},
		{
			name: "internal type",
			typ:  loadertest.Type(t, "internal/secret", "Token"),
			want: []props.Badge{
				{Label: "size", Value: "16 bytes"},
				{Label: "align", Value: "8"},
				{Label: "comparable"},
				{Label: "internal"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			badges := props.Badges(tc.typ)

			// Assert
			if got, want := badges, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Badges(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
