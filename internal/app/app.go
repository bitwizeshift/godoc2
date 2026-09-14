package app

import (
	"context"

	"github.com/bitwizeshift/go-cli"

	"github.com/bitwizeshift/godoc2/internal/generate"
	"github.com/bitwizeshift/godoc2/internal/loader"
)

// CommandID is the id path of the root command in the specification.
const CommandID = "godoc2"

// Generator runs a documentation build.
type Generator interface {
	Generate(ctx context.Context) error
}

// Builder collects the flag components and builds the [Runner].
type Builder struct {
	Output   *OutputFlags
	Progress *ProgressFlags
	Patterns *PatternArgs
}

// NewBuilder returns a [Builder] with fresh flag components.
func NewBuilder() *Builder {
	return &Builder{
		Output:   &OutputFlags{},
		Progress: &ProgressFlags{},
		Patterns: &PatternArgs{},
	}
}

// Build returns the [Runner] configured from the parsed arguments.
func (b *Builder) Build(ctx context.Context) (cli.Runner, error) {
	return &Runner{
		Generator: &generate.Generator{
			Loader: loader.Config{
				Patterns:  b.Patterns.Patterns(),
				Workspace: b.Patterns.Workspace(),
			},
			Sink:     b.Output.Sink(),
			Reporter: b.Progress.Reporter(ctx),
		},
	}, nil
}

var _ cli.Builder = (*Builder)(nil)

// Runner executes the generation.
type Runner struct {
	Generator Generator
}

// Run generates the documentation. It returns any error from the generator.
func (r *Runner) Run(ctx context.Context) error {
	return r.Generator.Generate(ctx)
}

var _ cli.Runner = (*Runner)(nil)
