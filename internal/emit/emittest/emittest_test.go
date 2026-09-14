package emittest_test

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/emit/emittest"
)

func TestMemSink_Create_StoresContent(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := emittest.NewMemSink()
	w, err := sut.Create("a/b.html")
	if err != nil {
		t.Fatalf("Create(...) = %v, want nil", err)
	}

	// Act
	_, writeErr := io.WriteString(w, "hello")
	closeErr := w.Close()
	content, ok := sut.Content("a/b.html")
	paths := sut.Paths()

	// Assert
	if got, want := writeErr, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("WriteString(...) = %v, want nil", got)
	}
	if got, want := closeErr, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Close() = %v, want nil", got)
	}
	if got, want := ok, true; !cmp.Equal(got, want) {
		t.Errorf("Content(...) ok = %v, want %v", got, want)
	}
	if got, want := content, "hello"; !cmp.Equal(got, want) {
		t.Errorf("Content(...) = %q, want %q", got, want)
	}
	if got, want := paths, []string{"a/b.html"}; !cmp.Equal(got, want) {
		t.Errorf("Paths() = %v, want %v", got, want)
	}
}

func TestErrSink_Create_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("boom")
	sut := emittest.ErrSink(testErr)

	// Act
	w, err := sut.Create("a.html")

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Create(...) = %v, want %v", got, want)
	}
	if got, want := w, (io.WriteCloser)(nil); !cmp.Equal(got, want) {
		t.Errorf("Create(...) writer = %v, want nil", got)
	}
}
