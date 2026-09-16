# `godoc2`

`godoc2` is a new, improved, and more comprehensive go documentation system.

This tool is inspired by `rustdoc`'s output format, and borrows a lot of those
ideas while modeling it around Go's idioms and language structure.

_Ask more of your go documentation_

## Quick Guide

* [Features](#features) \
  Why this tool is worth your time
* [Getting Started](#getting-started) \
  A quick step-by-step for how to use this tool
* [Example docs](https://rodusek.com/godoc2) \
  See the docs for _this_ tool as a proof-of-concept.

## Features

There are a variety of improvements that `godoc2` brings over the base godocs:

* Better navigation, such as comprehensive symbol searching
* True markdown support in comments, allowing for richer comment definitions
* Cross-package references for discoverability
  * Types indicate all constructors, even ones from different packages
  * Types indicate all interfaces they implement from within the build spec
  * Interfaces indicate all types that implement it from within the build spec
* Support for multi-module builds (native support for `go.work`)
* Native support for displaying a package's sentinel errors
* Support for documenting unexported types (for internal dev-docs)
* Types are classified by their properties (size, align, copyability, etc)

## Getting Started

### Installing

Installing `godoc2` is simple; just run:

```bash
go install github.com/bitwizeshift/godoc2
```

You can also, alternatively, install this as a _`tool`_ in your `go.mod` so that
it can be used as a `go tool` call:

```bash
go get -tool github.com/bitwizeshift/godoc2
```

which makes it possible to run this as `go tool godoc2`.

### Running

`godoc2` follows Go's methodology of keeping things "simple". To document
all packages under `./...`, just run `godoc2 ./...` -- and it'll create the
documentation in `dist/`.

To document an entire Go workspace from a `go.work`, you can use
`godoc2 --workspace go.work`.

See `godoc2 --help` for full details.

## License

This project is dual-licensed under both MIT and APACHE-2, at the user's
choice.
