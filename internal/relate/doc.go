// Package relate finds the relations between types, functions, values, and
// interfaces.
//
// An [Index] is built once per site. It answers which values are instances
// of a type, which functions construct or use it, which interfaces a type
// implements, and which types implement an interface. Interface and type
// candidates come from the modules of the site and from every package they
// import, except internal packages outside the site.
package relate
