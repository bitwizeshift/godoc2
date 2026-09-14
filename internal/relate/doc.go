// Package relate finds the relations between types, functions, values, and
// interfaces.
//
// An [Index] is built once per module. It answers which values are instances
// of a type, which functions construct or use it, which interfaces a type
// implements, and which types implement an interface. Interface and type
// candidates come from the module and from every package it imports, except
// internal packages of other modules.
package relate
