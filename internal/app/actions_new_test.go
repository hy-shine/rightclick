package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/locales"
)

func TestPasteAsFile(t *testing.T) {
	oldRead := readClipboardText
	oldNotify := notify
	t.Cleanup(func() {
		readClipboardText = oldRead
		notify = oldNotify
	})

	notifyCalled := false
	notify = func(title, body string) {
		notifyCalled = true
	}

	t.Run("EmptyClipboard", func(t *testing.T) {
		subDir := t.TempDir()
		readClipboardText = func(maxBytes int64) (string, error) { return "   \n\t", nil }
		err := pasteAsFile([]string{subDir})
		if err == nil || err.Error() != locales.ErrNoTextInClipboard() {
			t.Fatalf("expected ErrNoTextInClipboard, got %v", err)
		}
	})

	t.Run("OverSizeClipboard", func(t *testing.T) {
		subDir := t.TempDir()
		readClipboardText = func(maxBytes int64) (string, error) { return strings.Repeat("A", 1048577), nil }
		err := pasteAsFile([]string{subDir})
		if err == nil || err.Error() != locales.ErrClipboardTooLarge() {
			t.Fatalf("expected ErrClipboardTooLarge, got %v", err)
		}
	})

	t.Run("PlainTextCreation", func(t *testing.T) {
		subDir := t.TempDir()
		readClipboardText = func(maxBytes int64) (string, error) { return "Hello from clipboard", nil }
		notifyCalled = false
		err := pasteAsFile([]string{subDir})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(filepath.Join(subDir, "untitled.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "Hello from clipboard" {
			t.Errorf("got %q, want %q", string(data), "Hello from clipboard")
		}
		if !notifyCalled {
			t.Error("expected notification on success")
		}
	})

	t.Run("ExactBytesPreserved_NoTrailingNewline", func(t *testing.T) {
		subDir := t.TempDir()
		exact := "exact\nmultiline\ntext without trailing newline"
		readClipboardText = func(maxBytes int64) (string, error) { return exact, nil }
		err := pasteAsFile([]string{subDir})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(filepath.Join(subDir, "untitled.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != exact {
			t.Errorf("got bytes %q, want exact %q", string(data), exact)
		}
	})

	t.Run("JSONCreation", func(t *testing.T) {
		subDir := t.TempDir()
		readClipboardText = func(maxBytes int64) (string, error) { return `{"key": "value"}`, nil }
		err := pasteAsFile([]string{subDir})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(filepath.Join(subDir, "untitled.json"))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `{"key": "value"}` {
			t.Errorf("got %q, want %q", string(data), `{"key": "value"}`)
		}
	})

	t.Run("ReadOnlyDir", func(t *testing.T) {
		subDir := t.TempDir()
		readOnlyDir := filepath.Join(subDir, "readonly_dir")
		if err := os.Mkdir(readOnlyDir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0o755) })

		readClipboardText = func(maxBytes int64) (string, error) { return "some text", nil }
		err := pasteAsFile([]string{readOnlyDir})
		if err == nil || err.Error() != locales.ErrNoWritePermission() {
			t.Fatalf("expected ErrNoWritePermission, got %v", err)
		}
	})

	t.Run("NonexistentDir", func(t *testing.T) {
		subDir := t.TempDir()
		missingDir := filepath.Join(subDir, "missing_subdir")
		readClipboardText = func(maxBytes int64) (string, error) { return "some text", nil }
		err := pasteAsFile([]string{missingDir})
		if err == nil || err.Error() != locales.ErrFileNotFound() {
			t.Fatalf("expected ErrFileNotFound, got %v", err)
		}
	})
}

func TestCopyContentAction(t *testing.T) {
	dir := t.TempDir()

	oldCopy := copyText
	oldNotify := notify
	t.Cleanup(func() {
		copyText = oldCopy
		notify = oldNotify
	})

	var copied string
	copyText = func(text string) error {
		copied = text
		return nil
	}

	t.Run("NormalText", func(t *testing.T) {
		path := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(path, []byte("Content to copy"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := copyContent([]string{path}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if copied != "Content to copy" {
			t.Errorf("copied = %q, want %q", copied, "Content to copy")
		}
	})

	t.Run("EmptyFileNotifies", func(t *testing.T) {
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		notified := false
		notify = func(title, body string) {
			notified = true
		}
		if err := copyContent([]string{path}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if copied != "" {
			t.Errorf("copied = %q, want empty string", copied)
		}
		if !notified {
			t.Error("expected notification for empty file copy")
		}
	})

	t.Run("ReadOnlyFile", func(t *testing.T) {
		path := filepath.Join(dir, "unreadable.txt")
		if err := os.WriteFile(path, []byte("secret"), 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

		err := copyContent([]string{path})
		if err == nil || err.Error() != locales.ErrNoReadPermission() {
			t.Fatalf("expected ErrNoReadPermission, got %v", err)
		}
	})

	t.Run("NonexistentFile", func(t *testing.T) {
		missing := filepath.Join(dir, "missing.txt")
		err := copyContent([]string{missing})
		if err == nil || err.Error() != locales.ErrFileNotFound() {
			t.Fatalf("expected ErrFileNotFound, got %v", err)
		}
	})
}

func TestMakeExecutableAction(t *testing.T) {
	dir := t.TempDir()

	oldNotify := notify
	t.Cleanup(func() { notify = oldNotify })

	var notifiedBody string
	notify = func(title, body string) {
		notifiedBody = body
	}

	t.Run("SuccessAndAlreadyExecutable", func(t *testing.T) {
		f := filepath.Join(dir, "run.sh")
		if err := os.WriteFile(f, []byte("#!/bin/sh\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		// First run: should succeed and notify
		notifiedBody = ""
		if err := makeExecutableAction([]string{f}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if notifiedBody != locales.NotifyChmodSuccess(1) {
			t.Errorf("got notification %q, want %q", notifiedBody, locales.NotifyChmodSuccess(1))
		}

		// Second run: should recognize already executable
		notifiedBody = ""
		if err := makeExecutableAction([]string{f}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if notifiedBody != locales.NotifyChmodAlready() {
			t.Errorf("got notification %q, want %q", notifiedBody, locales.NotifyChmodAlready())
		}
	})

	t.Run("OnlyIgnoredTargets_NotifiesNoFiles", func(t *testing.T) {
		subDir := filepath.Join(dir, "sub_folder")
		if err := os.Mkdir(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		notifiedBody = ""
		if err := makeExecutableAction([]string{subDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if notifiedBody != locales.NotifyChmodNoFiles() {
			t.Errorf("got notification %q, want %q", notifiedBody, locales.NotifyChmodNoFiles())
		}
	})

	t.Run("PartialFailureReported", func(t *testing.T) {
		valid := filepath.Join(dir, "script_valid.sh")
		if err := os.WriteFile(valid, []byte("#!/bin/sh\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		missing := filepath.Join(dir, "missing_in_batch.sh")
		notifiedBody = ""
		if err := makeExecutableAction([]string{valid, missing}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(notifiedBody, "1") {
			t.Errorf("expected partial failure notification to mention counts, got: %q", notifiedBody)
		}
	})

	t.Run("FullFailureReported", func(t *testing.T) {
		missing := filepath.Join(dir, "nonexistent.sh")
		if err := makeExecutableAction([]string{missing}); err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestGetClipboardText(t *testing.T) {
	oldCmd := pbpasteCmd
	t.Cleanup(func() { pbpasteCmd = oldCmd })

	t.Run("ExactBytesPreserved_NoTrailingNewline", func(t *testing.T) {
		pbpasteCmd = func() *exec.Cmd {
			return exec.Command("/usr/bin/printf", "%s", "token_value_no_nl")
		}
		got, err := getClipboardText(1048576)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "token_value_no_nl" {
			t.Fatalf("got %q, want %q", got, "token_value_no_nl")
		}
	})

	t.Run("OversizeTruncation", func(t *testing.T) {
		pbpasteCmd = func() *exec.Cmd {
			return exec.Command("/usr/bin/printf", "%s", strings.Repeat("X", 200))
		}
		_, err := getClipboardText(50)
		if err == nil || err.Error() != locales.ErrClipboardTooLarge() {
			t.Fatalf("expected ErrClipboardTooLarge, got %v", err)
		}
	})
}
