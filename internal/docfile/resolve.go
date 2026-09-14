package docfile

import (
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
)

// Asset is a file inside the site that a documentation file links to. The
// generator copies it so the link stays valid.
type Asset struct {
	// Source is the file path on disk.
	Source string

	// Output is the path of the copy in the generated site.
	Output string
}

// Resolver maps the relative links of documentation files to the generated
// site.
type Resolver struct {
	site     *model.Site
	modules  map[string]*model.Module
	packages map[string]*model.Package
	sources  map[string]string
	assets   map[string]Asset
}

// NewResolver returns a [Resolver] for the modules of site.
func NewResolver(site *model.Site) *Resolver {
	r := &Resolver{
		site:     site,
		modules:  map[string]*model.Module{},
		packages: map[string]*model.Package{},
		sources:  map[string]string{},
		assets:   map[string]Asset{},
	}
	for _, mod := range site.Modules {
		r.modules[filepath.Clean(mod.Dir)] = mod
		for _, p := range mod.Packages {
			r.packages[filepath.Clean(p.Dir)] = p
			for _, f := range p.Files {
				r.sources[filepath.Clean(f.Path)] = pathmap.Source(mod.Path, p.RelPath, f.Name)
			}
		}
	}
	return r
}

// Resolve maps dest, a link destination written in f, to an href from the
// page at the output path from.
//
// A link to a module or package directory returns the href of its page. A
// link to a Go source file returns the href of its source page. A link to
// another file inside a module, or inside the site directory, returns the
// href of a copy of the file, which [Resolver.Assets] then lists. Any
// fragment is kept. Every other destination is returned unchanged.
func (r *Resolver) Resolve(f *model.DocFile, from, dest string) string {
	target, fragment, ok := splitDest(dest)
	if !ok {
		return dest
	}
	abs := filepath.Join(filepath.Dir(f.Path), filepath.FromSlash(target))
	if mod, ok := r.modules[abs]; ok {
		return pathmap.Rel(from, pathmap.Module(mod.Path)) + fragment
	}
	if p, ok := r.packages[abs]; ok {
		return pathmap.Rel(from, pathmap.Package(p.Module.Path, p.RelPath)) + fragment
	}
	if page, ok := r.sources[abs]; ok {
		return pathmap.Rel(from, page) + fragment
	}
	output, ok := r.copyPath(abs)
	if !ok || !isRegularFile(abs) {
		return dest
	}
	r.assets[output] = Asset{Source: abs, Output: output}
	return pathmap.Rel(from, output) + fragment
}

// Assets lists the files that resolved links point at, sorted by output
// path.
func (r *Resolver) Assets() []Asset {
	assets := slices.Collect(maps.Values(r.assets))
	slices.SortFunc(assets, func(lhs, rhs Asset) int {
		return strings.Compare(lhs.Output, rhs.Output)
	})
	return assets
}

// splitDest separates a relative destination into its path and fragment. It
// reports false for destinations that are not relative paths: URLs with a
// scheme or host, absolute paths, and fragment-only links.
func splitDest(dest string) (target, fragment string, ok bool) {
	u, err := url.Parse(dest)
	if err != nil || u.IsAbs() || u.Host != "" || u.Path == "" || strings.HasPrefix(u.Path, "/") {
		return "", "", false
	}
	if u.Fragment != "" {
		fragment = "#" + u.EscapedFragment()
	}
	return u.Path, fragment, true
}

// copyPath returns the output path of a copy of the file at abs: below the
// module directory of the innermost module that holds it, or else below the
// output root when the site directory holds it. It reports false for a file
// outside both.
func (r *Resolver) copyPath(abs string) (string, bool) {
	if mod := r.owner(abs); mod != nil {
		rel, ok := relativeTo(mod.Dir, abs)
		return pathmap.File(mod.Path, rel), ok
	}
	rel, ok := relativeTo(r.site.Dir, abs)
	return pathmap.File("", rel), ok
}

// owner returns the innermost module whose directory holds abs, or nil.
func (r *Resolver) owner(abs string) *model.Module {
	var owner *model.Module
	for _, mod := range r.site.Modules {
		if _, ok := relativeTo(mod.Dir, abs); ok && (owner == nil || len(mod.Dir) > len(owner.Dir)) {
			owner = mod
		}
	}
	return owner
}

// relativeTo returns the slash-separated path of abs relative to dir, or
// false when abs lies outside dir.
func relativeTo(dir, abs string) (string, bool) {
	rel, err := filepath.Rel(dir, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if rel == "." {
		return "", true
	}
	return filepath.ToSlash(rel), true
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
