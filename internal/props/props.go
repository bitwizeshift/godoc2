package props

import (
	"go/token"
	"go/types"
	"strconv"

	"github.com/bitwizeshift/godoc2/internal/model"
)

// LargeThreshold is the size in bytes above which a type is marked large.
const LargeThreshold = 80

// Badge is one property of a type. Value is empty for boolean properties.
type Badge struct {
	Label string
	Value string
}

// titles is the hover text of every badge label.
var titles = map[string]string{
	"size":        "Size in bytes on a 64-bit system.",
	"align":       "Alignment in bytes on a 64-bit system.",
	"large":       "Size is above " + strconv.Itoa(LargeThreshold) + " bytes. Construct and pass by pointer to avoid large copies.",
	"comparable":  "Values can be compared with == and !=, and can be map keys.",
	"ordered":     "Values can be compared with <, <=, >, and >=.",
	"sealed":      "Interface with an unexported method. Only types in the same package can implement it.",
	"noncopiable": "Must not be copied. It holds data that requires object identity, such as a lock.",
	"internal":    "Importable only by packages rooted at the parent of the internal directory.",
	"embedded":    "Embedded in the struct. Its fields and methods are promoted to the struct.",
	"unexported":  "Not exported. It cannot be named outside its package.",
}

// Title returns the hover text that explains the badge with label. It returns
// an empty string for a label that has no badge.
func Title(label string) string {
	return titles[label]
}

var sizes = types.SizesFor("gc", "amd64")

// Badges returns the properties of t in display order: size, alignment,
// large, comparable, ordered, sealed, noncopiable, and internal.
func Badges(t *model.Type) []Badge {
	if t.Obj == nil {
		return nil
	}
	typ := types.Unalias(t.Obj.Type())
	var badges []Badge
	if size, align, ok := layout(typ); ok {
		badges = append(badges,
			Badge{Label: "size", Value: strconv.FormatInt(size, 10) + " bytes"},
			Badge{Label: "align", Value: strconv.FormatInt(align, 10)},
		)
		if size > LargeThreshold {
			badges = append(badges, Badge{Label: "large"})
		}
	}
	if types.Comparable(typ) {
		badges = append(badges, Badge{Label: "comparable"})
	}
	if ordered(typ) {
		badges = append(badges, Badge{Label: "ordered"})
	}
	if sealed(typ) {
		badges = append(badges, Badge{Label: "sealed"})
	}
	if noncopiable(typ, map[types.Type]bool{}) {
		badges = append(badges, Badge{Label: "noncopiable"})
	}
	if t.Pkg != nil && t.Pkg.Internal() {
		badges = append(badges, Badge{Label: "internal"})
	}
	return badges
}

// layout returns the size and alignment of typ. Types whose size depends on
// type parameters report false.
func layout(typ types.Type) (size, align int64, ok bool) {
	if hasTypeParams(typ) {
		return 0, 0, false
	}
	return sizes.Sizeof(typ), sizes.Alignof(typ), true
}

// hasTypeParams reports whether typ is a generic type that has not been
// instantiated.
func hasTypeParams(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	return ok && named.TypeParams().Len() > 0 && named.TypeArgs().Len() == 0
}

// lockerType is the method set that marks a lock: sync.Locker. It is
// completed once here because [types.Implements] fills in the type set of an
// incomplete interface on first use, which is not safe from several
// goroutines.
var lockerType = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "Lock", types.NewSignatureType(nil, nil, nil, nil, nil, false)),
	types.NewFunc(token.NoPos, nil, "Unlock", types.NewSignatureType(nil, nil, nil, nil, nil, false)),
}, nil).Complete()

// noncopiable reports whether typ must not be copied, with the rule of the
// go vet copylocks check: a struct whose pointer type is a [sync.Locker]
// while its value type is not, the sync.noCopy marker, or a struct or array
// that holds such a type.
func noncopiable(typ types.Type, seen map[types.Type]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true
	for {
		arr, ok := typ.Underlying().(*types.Array)
		if !ok {
			break
		}
		typ = arr.Elem()
	}
	st, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	if types.Implements(types.NewPointer(typ), lockerType) && !types.Implements(typ, lockerType) {
		return true
	}
	if named, ok := typ.(*types.Named); ok && named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "noCopy" {
		return true
	}
	for f := range st.Fields() {
		if noncopiable(f.Type(), seen) {
			return true
		}
	}
	return false
}

// ordered reports whether typ supports the <, <=, >, and >= operators: its
// underlying type is an integer, a float, or a string.
func ordered(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsOrdered != 0
}

// sealed reports whether typ is an interface with an unexported method.
func sealed(typ types.Type) bool {
	iface, ok := typ.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	for m := range iface.Methods() {
		if !m.Exported() {
			return true
		}
	}
	return false
}
