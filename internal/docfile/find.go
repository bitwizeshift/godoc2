package docfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bitwizeshift/godoc2/internal/model"
)

// Names are the file names that document a directory, in precedence order.
var Names = []string{"index.md", "README.md"}

// Find returns the documentation file of dir, or nil when dir holds none of
// [Names]. It returns an error when a candidate file cannot be read.
func Find(dir string) (*model.DocFile, error) {
	for _, name := range Names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("docfile: %w", err)
		}
		return &model.DocFile{Path: path, Text: string(data)}, nil
	}
	return nil, nil
}
