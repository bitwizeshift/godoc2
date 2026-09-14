// Command godoc2 generates rustdoc-style HTML documentation for a Go module.
package main

import (
	_ "embed"

	"github.com/bitwizeshift/go-cli"

	"github.com/bitwizeshift/godoc2/internal/app"
)

//go:embed app.yaml
var appYAML []byte

func main() {
	cli.FromBytes(appYAML,
		cli.BindBuilder(app.CommandID, app.NewBuilder()),
	).Execute()
}
