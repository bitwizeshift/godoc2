package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/app"
)

// errGenerator is an [app.Generator] that always fails.
type errGenerator struct {
	err error
}

func (g errGenerator) Generate(context.Context) error {
	return g.err
}

// noOpGenerator is an [app.Generator] that does nothing.
type noOpGenerator struct{}

func (noOpGenerator) Generate(context.Context) error {
	return nil
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	testErr := errors.New("boom")

	testCases := []struct {
		name      string
		generator app.Generator
		wantErr   error
	}{
		{
			name:      "success",
			generator: noOpGenerator{},
			wantErr:   nil,
		},
		{
			name:      "generator fails",
			generator: errGenerator{err: testErr},
			wantErr:   testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			sut := &app.Runner{Generator: tc.generator}

			// Act
			err := sut.Run(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Runner.Run(...) = %v, want %v", got, want)
			}
		})
	}
}

func TestBuilder_Build_ReturnsRunner(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := app.NewBuilder()

	// Act
	runner, err := sut.Build(ctx)
	_, isRunner := runner.(*app.Runner)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Builder.Build(...) = %v, want nil", got)
	}
	if got, want := isRunner, true; !cmp.Equal(got, want) {
		t.Errorf("Builder.Build(...) returns *app.Runner = %v, want %v", got, want)
	}
}
