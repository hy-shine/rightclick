package files

import (
	"slices"
	"testing"
)

func TestOpenArgs(t *testing.T) {
	got := openArgs("Visual Studio Code", "/tmp/a b.txt")
	want := []string{"-a", "Visual Studio Code", "/tmp/a b.txt"}
	if !slices.Equal(got, want) {
		t.Errorf("openArgs = %q, want %q", got, want)
	}
}
