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

// fixture loads one fixture module once per test binary.
type fixture struct {
	dir  string
	once sync.Once
	mod  *model.Module
	err  error
}

func (f *fixture) load(t testing.TB) *model.Module {
	t.Helper()
	f.once.Do(func() {
		f.mod, f.err = loader.Load(context.Background(), loader.Config{
			Dir:      f.dir,
			Patterns: []string{"./..."},
		})
	})
	if f.err != nil {
		t.Fatalf("loader.Load(...) = %v, want nil", f.err)
	}
	return f.mod
}

var (
	sample = &fixture{dir: fixtureDir("sample")}
	bare   = &fixture{dir: fixtureDir("bare")}
)

// Sample returns the loaded fixture module. The module is loaded once per
// test binary and shared, so callers must not modify it.
func Sample(t testing.TB) *model.Module {
	t.Helper()
	return sample.load(t)
}

// SampleDir returns the absolute directory of the fixture module.
func SampleDir() string {
	return sample.dir
}

// Bare returns the loaded fixture module that has a README.md and no package
// in its root directory. The module is loaded once per test binary and
// shared, so callers must not modify it.
func Bare(t testing.TB) *model.Module {
	t.Helper()
	return bare.load(t)
}

// BareDir returns the absolute directory of the bare fixture module.
func BareDir() string {
	return bare.dir
}

func fixtureDir(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "testdata", name)
}

// Package returns the fixture package with the given path relative to the
// module root. It fails the test when the package does not exist.
func Package(t testing.TB, rel string) *model.Package {
	t.Helper()
	for _, p := range Sample(t).Packages {
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
