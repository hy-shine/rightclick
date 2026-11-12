package finder

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// fakeCommands replaces runCommand for one test. Every command gets the same
// output and error. It returns a pointer to the list of commands that ran.
func fakeCommands(t *testing.T, out string, err error) *[]string {
	t.Helper()
	var ran []string
	old := runCommand
	runCommand = func(name string, args ...string) ([]byte, error) {
		ran = append(ran, strings.Join(append([]string{name}, args...), " "))
		return []byte(out), err
	}
	t.Cleanup(func() { runCommand = old })
	return &ran
}

func TestHiddenFilesShown(t *testing.T) {
	missing := "The domain/default pair of (com.apple.finder, AppleShowAllFiles) does not exist"
	exitErr := errors.New("exit status 1")
	tests := []struct {
		name    string
		out     string
		err     error
		want    bool
		wantErr bool
	}{
		{name: "shown", out: "1\n", want: true},
		{name: "hidden", out: "0\n", want: false},
		{name: "key missing", out: missing, err: exitErr, want: false},
		{name: "other failure", out: "defaults: cannot read", err: exitErr, wantErr: true},
		{name: "unexpected value", out: "YES\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCommands(t, tt.out, tt.err)
			got, err := HiddenFilesShown()
			if (err != nil) != tt.wantErr {
				t.Fatalf("HiddenFilesShown() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("HiddenFilesShown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetHiddenFilesShownWritesThenRestarts(t *testing.T) {
	tests := []struct {
		show  bool
		value string
	}{
		{show: true, value: "TRUE"},
		{show: false, value: "FALSE"},
	}
	for _, tt := range tests {
		ran := fakeCommands(t, "", nil)
		if err := SetHiddenFilesShown(tt.show); err != nil {
			t.Fatal(err)
		}
		want := []string{
			"defaults write com.apple.finder AppleShowAllFiles -bool " + tt.value,
			"killall Finder",
		}
		if !slices.Equal(*ran, want) {
			t.Errorf("SetHiddenFilesShown(%v) ran %q, want %q", tt.show, *ran, want)
		}
	}
}

func TestFailedWriteDoesNotRestartFinder(t *testing.T) {
	ran := fakeCommands(t, "permission denied", errors.New("exit status 1"))
	if err := SetHiddenFilesShown(true); err == nil {
		t.Fatal("expected an error when defaults write fails")
	}
	if len(*ran) != 1 {
		t.Errorf("commands %q, want only the write", *ran)
	}
}
