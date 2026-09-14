package docfile_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
)

func TestFind(t *testing.T) {
	t.Parallel()

	sample := loadertest.SampleDir()

	testCases := []struct {
		name string
		dir  string
		want *model.DocFile
	}{
		{
			name: "no file",
			dir:  filepath.Join(sample, "empty"),
			want: nil,
		},
		{
			name: "index and README",
			dir:  filepath.Join(sample, "indexed"),
			want: &model.DocFile{
				Path: filepath.Join(sample, "indexed", "index.md"),
				Text: "# Indexed\n\nPackage indexed is documented by index.md.\n",
			},
		},
		{
			name: "README only",
			dir:  filepath.Join(sample, "shapes"),
			want: &model.DocFile{
				Path: filepath.Join(sample, "shapes", "README.md"),
				Text: "# Shapes README\n\nThis file is not used because the package has a doc comment.\n",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			f, err := docfile.Find(tc.dir)

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Find(...) = %v, want nil", got)
			}
			if got, want := f, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Find(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

