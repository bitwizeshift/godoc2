package app_test

import (
	"context"
	"testing"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/clitest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/app"
)

func TestBuilder_RegistersArguments(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	sut := app.NewBuilder()
	wantFlags := []*argtest.Flag{
		{Long: "output", Shorthand: "o", ValueLabel: "dir", Group: "Output Flags"},
		{Long: "verbose", Shorthand: "v", ValueLabel: "bool", Group: "Output Flags"},
		{Long: "workspace", Shorthand: "w", ValueLabel: "file", Group: "Input Flags"},
	}
	wantUnmatched := &argtest.Unmatched{
		Name:  "patterns",
		Usage: "packages to document, as accepted by go build (default ./...)",
	}

	// Act
	arg.Register(cl, sut)
	flags := argtest.AllFlags(cl)
	unmatched := argtest.GetUnmatched(cl)

	// Assert
	if got, want := flags, wantFlags; !cmp.Equal(got, want, cmpopts.IgnoreFields(argtest.Flag{}, "ExclusiveWith", "RequiredWith", "OneRequiredWith")) {
		t.Errorf("AllFlags(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := unmatched, wantUnmatched; !cmp.Equal(got, want) {
		t.Errorf("GetUnmatched(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestOutputFlags_Dir(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "default",
			args: nil,
			want: "dist",
		},
		{
			name: "long flag",
			args: []string{"--output", "docs"},
			want: "docs",
		},
		{
			name: "short flag",
			args: []string{"-o", "public"},
			want: "public",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			sut := &app.OutputFlags{}
			arg.Register(cl, sut)
			argtest.MustParse(t, cl, tc.args...)

			// Act
			dir := sut.Dir()

			// Assert
			if got, want := dir, tc.want; !cmp.Equal(got, want) {
				t.Errorf("OutputFlags.Dir() = %q, want %q", got, want)
			}
		})
	}
}

func TestProgressFlags_Verbose(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "default",
			args: nil,
			want: false,
		},
		{
			name: "verbose",
			args: []string{"-v"},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			sut := &app.ProgressFlags{}
			arg.Register(cl, sut)
			argtest.MustParse(t, cl, tc.args...)

			// Act
			verbose := sut.Verbose()

			// Assert
			if got, want := verbose, tc.want; !cmp.Equal(got, want) {
				t.Errorf("ProgressFlags.Verbose() = %v, want %v", got, want)
			}
		})
	}
}

func TestProgressFlags_Reporter_WritesToOutStream(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, output := clitest.WithCaptureWriters(context.Background())
	sut := &app.ProgressFlags{}

	// Act
	reporter := sut.Reporter(ctx)
	reporter.Stage("Loading packages")
	reporter.File("a.html")
	text := output.Stdout.String()

	// Assert
	if got, want := text, "Loading packages\n"; !cmp.Equal(got, want) {
		t.Errorf("Reporter output = %q, want %q", got, want)
	}
}

func TestPatternArgs_Patterns(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "default",
			args: nil,
			want: []string{"./..."},
		},
		{
			name: "explicit patterns",
			args: []string{"./cmd/...", "./internal/app"},
			want: []string{"./cmd/...", "./internal/app"},
		},
		{
			name: "workspace only",
			args: []string{"--workspace", "go.work"},
			want: nil,
		},
		{
			name: "workspace and patterns",
			args: []string{"-w", "go.work", "./cmd/..."},
			want: []string{"./cmd/..."},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			sut := &app.PatternArgs{}
			arg.Register(cl, sut)
			argtest.MustParse(t, cl, tc.args...)

			// Act
			patterns := sut.Patterns()

			// Assert
			if got, want := patterns, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("PatternArgs.Patterns() = %v, want %v", got, want)
			}
		})
	}
}

func TestPatternArgs_Workspace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "default",
			args: nil,
			want: "",
		},
		{
			name: "long flag",
			args: []string{"--workspace", "go.work"},
			want: "go.work",
		},
		{
			name: "short flag",
			args: []string{"-w", "tools/go.work"},
			want: "tools/go.work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			sut := &app.PatternArgs{}
			arg.Register(cl, sut)
			argtest.MustParse(t, cl, tc.args...)

			// Act
			workspace := sut.Workspace()

			// Assert
			if got, want := workspace, tc.want; !cmp.Equal(got, want) {
				t.Errorf("PatternArgs.Workspace() = %q, want %q", got, want)
			}
		})
	}
}
