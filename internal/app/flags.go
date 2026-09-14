package app

import (
	"context"

	cli "github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/arg"

	"github.com/bitwizeshift/godoc2/internal/emit"
	"github.com/bitwizeshift/godoc2/internal/progress"
)

const (
	// DefaultOutput is the output directory used when --output is not given.
	DefaultOutput = "dist"

	// DefaultPattern is the package pattern used when none is given.
	DefaultPattern = "./..."

	outputGroup = "Output Flags"
)

// OutputFlags owns the --output flag and builds the output sink from it.
type OutputFlags struct {
	output string
}

// RegisterArgs registers --output/-o.
func (f *OutputFlags) RegisterArgs(cl *arg.CommandLine) {
	output := arg.Flag("output", &f.output,
		arg.Shorthand("o"),
		arg.ValueLabel("dir"),
		arg.Usage("directory to write the documentation site to"),
		arg.DefaultValue(DefaultOutput),
		arg.CompleteDirs(),
	)
	cl.Add(output)
	arg.Group(outputGroup, output)
}

var _ arg.Registrar = (*OutputFlags)(nil)

// Dir returns the output directory.
func (f *OutputFlags) Dir() string {
	if f.output == "" {
		return DefaultOutput
	}
	return f.output
}

// Sink returns the sink that writes into the output directory.
func (f *OutputFlags) Sink() emit.Sink {
	return emit.NewDirSink(f.Dir())
}

// ProgressFlags owns the --verbose flag and builds the progress reporter.
type ProgressFlags struct {
	verbose bool
}

// RegisterArgs registers --verbose/-v.
func (f *ProgressFlags) RegisterArgs(cl *arg.CommandLine) {
	verbose := arg.Flag("verbose", &f.verbose,
		arg.Shorthand("v"),
		arg.Usage("print every generated file, not only the stages"),
	)
	cl.Add(verbose)
	arg.Group(outputGroup, verbose)
}

var _ arg.Registrar = (*ProgressFlags)(nil)

// Verbose reports whether file names are printed.
func (f *ProgressFlags) Verbose() bool {
	return f.verbose
}

// Reporter returns a reporter that prints to the output stream of ctx.
func (f *ProgressFlags) Reporter(ctx context.Context) progress.Reporter {
	return progress.NewWriter(cli.OutStream(ctx), f.verbose)
}

// PatternArgs owns the package pattern arguments.
type PatternArgs struct {
	patterns []string
}

// RegisterArgs registers the trailing pattern arguments.
func (p *PatternArgs) RegisterArgs(cl *arg.CommandLine) {
	cl.Add(arg.Unmatched("patterns", &p.patterns,
		arg.Usage("packages to document, as accepted by go build (default ./...)"),
		arg.DefaultValue(DefaultPattern),
		arg.CompleteDirs(),
	))
}

var _ arg.Registrar = (*PatternArgs)(nil)

// Patterns returns the package patterns.
func (p *PatternArgs) Patterns() []string {
	if len(p.patterns) == 0 {
		return []string{DefaultPattern}
	}
	return p.patterns
}
