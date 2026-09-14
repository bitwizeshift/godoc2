package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

// workspace returns the patterns and environment of a load. Without a
// workspace file both are the configured patterns and a nil environment,
// which lets the go command use the environment of the process.
//
// With a workspace file, the modules the file uses come first as "<dir>/..."
// patterns relative to dir, and GOWORK points the go command at the file.
func workspace(dir string, cfg Config) (patterns []string, env []string, err error) {
	if cfg.Workspace == "" {
		return cfg.Patterns, nil, nil
	}
	file := cfg.Workspace
	if !filepath.IsAbs(file) {
		file = filepath.Join(dir, file)
	}
	dirs, err := workspaceDirs(file)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrLoad, err)
	}
	for _, d := range dirs {
		patterns = append(patterns, localPattern(dir, d))
	}
	patterns = append(patterns, cfg.Patterns...)
	env = append(os.Environ(), "GOWORK="+file)
	return patterns, env, nil
}

// workspaceDirs returns the absolute module directories that the go.work
// file at path uses, in file order.
func workspaceDirs(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	wf, err := modfile.ParseWork(path, data, nil)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, use := range wf.Use {
		d := use.Path
		if !filepath.IsAbs(d) {
			d = filepath.Join(filepath.Dir(path), d)
		}
		dirs = append(dirs, filepath.Clean(d))
	}
	return slices.Compact(dirs), nil
}

// localPattern returns the "..." pattern that matches every package below
// target, written relative to dir when target lies inside it.
func localPattern(dir, target string) string {
	rel, err := filepath.Rel(dir, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.Join(target, "...")
	}
	return "./" + filepath.ToSlash(filepath.Join(rel, "..."))
}
