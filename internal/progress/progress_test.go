package progress_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/progress"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		verbose bool
		want    string
	}{
		{
			name:    "quiet",
			verbose: false,
			want:    "Loading\n",
		},
		{
			name:    "verbose",
			verbose: true,
			want:    "Loading\n  a/b.html\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var out strings.Builder
			sut := progress.NewWriter(&out, tc.verbose)

			// Act
			sut.Stage("Loading")
			sut.File("a/b.html")
			text := out.String()

			// Assert
			if got, want := text, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Writer output = %q, want %q", got, want)
			}
		})
	}
}
