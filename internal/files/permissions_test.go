package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveTemplatesIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "support", "templates.json")
	if err := SaveTemplates(path, []Template{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("templates file mode = %v, want 0600", got)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("templates folder mode = %v, want 0700", got)
	}
}

func TestMakeExecutable(t *testing.T) {
	dir := t.TempDir()
	regFile := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(regFile, []byte("#!/bin/sh\necho hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(dir, "link.sh")
	if err := os.Symlink(regFile, symlink); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.sh")

	// 1. First run: regFile should succeed; symlink and subDir ignored; missing failed.
	res := MakeExecutable([]string{regFile, symlink, subDir, missing})
	if len(res.Succeeded) != 1 || res.Succeeded[0] != regFile {
		t.Errorf("expected 1 succeeded (%s), got: %v", regFile, res.Succeeded)
	}
	if len(res.Ignored) != 2 {
		t.Errorf("expected 2 ignored (symlink and dir), got: %v", res.Ignored)
	}
	if len(res.Failed) != 1 || res.Failed[0].Path != missing {
		t.Errorf("expected 1 failed (missing), got: %v", res.Failed)
	}

	info, err := os.Lstat(regFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 != 0o111 {
		t.Errorf("expected executable bit set, got: %v", info.Mode().Perm())
	}

	// 2. Second run: regFile is already executable, should be in AlreadyExecutable.
	res2 := MakeExecutable([]string{regFile})
	if len(res2.Succeeded) != 0 {
		t.Errorf("expected 0 succeeded, got: %v", res2.Succeeded)
	}
	if len(res2.AlreadyExecutable) != 1 || res2.AlreadyExecutable[0] != regFile {
		t.Errorf("expected 1 already executable, got: %v", res2.AlreadyExecutable)
	}

	// 3. Special permission bits (ModeSticky) should be preserved.
	stickyFile := filepath.Join(dir, "sticky.sh")
	if err := os.WriteFile(stickyFile, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(stickyFile, os.ModeSticky|0o644); err != nil {
		t.Fatal(err)
	}
	resSticky := MakeExecutable([]string{stickyFile})
	if len(resSticky.Succeeded) != 1 {
		t.Errorf("expected stickyFile to succeed, got: %v", resSticky.Succeeded)
	}
	stickyInfo, err := os.Lstat(stickyFile)
	if err != nil {
		t.Fatal(err)
	}
	if stickyInfo.Mode()&os.ModeSticky == 0 {
		t.Errorf("expected ModeSticky to be preserved, got mode: %v", stickyInfo.Mode())
	}
	if stickyInfo.Mode().Perm()&0o111 != 0o111 {
		t.Errorf("expected executable bit to be set on stickyFile, got mode: %v", stickyInfo.Mode().Perm())
	}

	// 4. Batch partial failure: some files succeed while others fail.
	validBatch := filepath.Join(dir, "valid_batch.sh")
	if err := os.WriteFile(validBatch, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	batchRes := MakeExecutable([]string{validBatch, filepath.Join(dir, "bad1.sh"), filepath.Join(dir, "bad2.sh")})
	if len(batchRes.Succeeded) != 1 || batchRes.Succeeded[0] != validBatch {
		t.Errorf("expected 1 succeeded in batch, got: %v", batchRes.Succeeded)
	}
	if len(batchRes.Failed) != 2 {
		t.Errorf("expected 2 failed in batch, got: %v", batchRes.Failed)
	}
}
