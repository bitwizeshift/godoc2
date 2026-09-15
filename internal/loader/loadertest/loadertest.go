package loadertest

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/bitwizeshift/godoc2/internal/loader"
	"github.com/bitwizeshift/godoc2/internal/model"
)

// SamplePath is the module path of the fixture module.
const SamplePath = "example.com/sample"

// BarePath is the module path of the fixture module without a root package.
const BarePath = "example.com/bare"

// Module paths of the multi fixture: a root module with two nested modules
// in one go.work workspace.
const (
	MultiPath      = "example.com/multi"
	MultiAlphaPath = "example.com/multi/alpha"
	MultiBetaPath  = "example.com/multi/beta"
)

// fixture loads one fixture site once per test binary.
type fixture struct {
	dir        string
	workspace  string
	unexported bool
	once       sync.Once
	site       *model.Site
	err        error
}

func (f *fixture) load(t testing.TB) *model.Site {
	t.Helper()
	f.once.Do(func() {
		cfg := loader.Config{Dir: f.dir, Workspace: f.workspace, Unexported: f.unexported}
		if f.workspace == "" {
			cfg.Patterns = []string{"./..."}
		}
		f.site, f.err = loader.Load(context.Background(), cfg)
	})
	if f.err != nil {
		t.Fatalf("loader.Load(...) = %v, want nil", f.err)
	}
	return f.site
}

var (
	sample     = &fixture{dir: fixtureDir("sample")}
	unexported = &fixture{dir: fixtureDir("sample"), unexported: true}
	bare       = &fixture{dir: fixtureDir("bare")}
	multi      = &fixture{dir: fixtureDir("multi"), workspace: "go.work"}
)

// Sample returns the loaded fixture site, which holds the sample module. The
// site is loaded once per test binary and shared, so callers must not modify
// it.
func Sample(t testing.TB) *model.Site {
	t.Helper()
	return sample.load(t)
}

// Unexported returns the sample fixture site loaded with its unexported
// symbols. The site is loaded once per test binary and shared, so callers
// must not modify it.
func Unexported(t testing.TB) *model.Site {
	t.Helper()
	return unexported.load(t)
}

// UnexportedPackage returns the package at rel inside the sample module of
// the site that holds unexported symbols. It fails the test when the package
// does not exist.
func UnexportedPackage(t testing.TB, rel string) *model.Package {
	t.Helper()
	for _, p := range Unexported(t).Modules[0].Packages {
		if p.RelPath == rel {
			return p
		}
	}
	t.Fatalf("UnexportedPackage(%q): not found in fixture", rel)
	return nil
}

// UnexportedType returns the named type from the package at rel of the site
// that holds unexported symbols. It fails the test when the type does not
// exist.
func UnexportedType(t testing.TB, rel, name string) *model.Type {
	t.Helper()
	for _, typ := range UnexportedPackage(t, rel).Types {
		if typ.Name == name {
			return typ
		}
	}
	t.Fatalf("UnexportedType(%q, %q): not found in fixture", rel, name)
	return nil
}

// SampleModule returns the sample module of the fixture site.
func SampleModule(t testing.TB) *model.Module {
	t.Helper()
	return Sample(t).Modules[0]
}

// SampleDir returns the absolute directory of the fixture module.
func SampleDir() string {
	return sample.dir
}

// Bare returns the loaded fixture site of the module that has a README.md
// and no package in its root directory. The site is loaded once per test
// binary and shared, so callers must not modify it.
func Bare(t testing.TB) *model.Site {
	t.Helper()
	return bare.load(t)
}

// BareDir returns the absolute directory of the bare fixture module.
func BareDir() string {
	return bare.dir
}

// Multi returns the loaded fixture site of the go.work workspace with three
// modules. The site is loaded once per test binary and shared, so callers
// must not modify it.
func Multi(t testing.TB) *model.Site {
	t.Helper()
	return multi.load(t)
}

// MultiDir returns the absolute directory of the multi fixture workspace.
func MultiDir() string {
	return multi.dir
}

// MultiPackage returns the package at rel inside the multi fixture module
// at modulePath. It fails the test when the package does not exist.
func MultiPackage(t testing.TB, modulePath, rel string) *model.Package {
	t.Helper()
	if mod := Multi(t).Module(modulePath); mod != nil {
		for _, p := range mod.Packages {
			if p.RelPath == rel {
				return p
			}
		}
	}
	t.Fatalf("MultiPackage(%q, %q): not found in fixture", modulePath, rel)
	return nil
}

// MultiType returns the named type from the multi fixture package at rel
// inside the module at modulePath. It fails the test when the type does not
// exist.
func MultiType(t testing.TB, modulePath, rel, name string) *model.Type {
	t.Helper()
	for _, typ := range MultiPackage(t, modulePath, rel).Types {
		if typ.Name == name {
			return typ
		}
	}
	t.Fatalf("MultiType(%q, %q, %q): not found in fixture", modulePath, rel, name)
	return nil
}

func fixtureDir(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "testdata", name)
}

// Package returns the sample package with the given path relative to the
// module root. It fails the test when the package does not exist.
func Package(t testing.TB, rel string) *model.Package {
	t.Helper()
	for _, p := range SampleModule(t).Packages {
		if p.RelPath == rel {
			return p
		}
	}
	t.Fatalf("Package(%q): not found in fixture", rel)
	return nil
}

// Type returns the named type from the fixture package at rel. It fails the
// test when the type does not exist.
func Type(t testing.TB, rel, name string) *model.Type {
	t.Helper()
	for _, typ := range Package(t, rel).Types {
		if typ.Name == name {
			return typ
		}
	}
	t.Fatalf("Type(%q, %q): not found in fixture", rel, name)
	return nil
}

// Func returns the named function from the fixture package at rel. It fails
// the test when the function does not exist.
func Func(t testing.TB, rel, name string) *model.Func {
	t.Helper()
	for _, f := range Package(t, rel).Funcs {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("Func(%q, %q): not found in fixture", rel, name)
	return nil
}

// Method returns the named method of a fixture type. It fails the test when
// the method does not exist.
func Method(t testing.TB, rel, typeName, name string) *model.Func {
	t.Helper()
	for _, m := range Type(t, rel, typeName).Methods {
		if m.Name == name {
			return m
		}
	}
	t.Fatalf("Method(%q, %q, %q): not found in fixture", rel, typeName, name)
	return nil
}

// Field returns the named field of a fixture struct type. It fails the test
// when the field does not exist.
func Field(t testing.TB, rel, typeName, name string) *model.Field {
	t.Helper()
	for _, f := range Type(t, rel, typeName).Fields {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("Field(%q, %q, %q): not found in fixture", rel, typeName, name)
	return nil
}

// Value returns the named constant or variable from the fixture package at
// rel. It fails the test when the value does not exist.
func Value(t testing.TB, rel, name string) *model.Value {
	t.Helper()
	p := Package(t, rel)
	for _, v := range p.Consts {
		if v.Name == name {
			return v
		}
	}
	for _, v := range p.Vars {
		if v.Name == name {
			return v
		}
	}
	t.Fatalf("Value(%q, %q): not found in fixture", rel, name)
	return nil
}
