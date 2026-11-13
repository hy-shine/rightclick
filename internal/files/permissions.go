package files

import (
	"os"
)

// FileFailure records a file path and the error that prevented it from being updated.
type FileFailure struct {
	Path string
	Err  error
}

// ExecutableResult holds the partitioned outcome of MakeExecutable.
type ExecutableResult struct {
	Succeeded         []string
	AlreadyExecutable []string
	Ignored           []string
	Failed            []FileFailure
}

// MakeExecutable adds execution permissions (| 0o111) to regular files in paths.
// It skips non-regular files and symbolic links (Ignored), and files that are already
// executable (AlreadyExecutable). It preserves special permission bits (setuid, setgid, sticky).
func MakeExecutable(paths []string) ExecutableResult {
	const specialBits = os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky
	var result ExecutableResult
	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			result.Failed = append(result.Failed, FileFailure{Path: p, Err: err})
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			result.Ignored = append(result.Ignored, p)
			continue
		}
		if info.Mode().Perm()&0o111 == 0o111 {
			result.AlreadyExecutable = append(result.AlreadyExecutable, p)
			continue
		}
		newMode := (info.Mode() & specialBits) | 0o111
		if err := os.Chmod(p, newMode); err != nil {
			result.Failed = append(result.Failed, FileFailure{Path: p, Err: err})
			continue
		}
		result.Succeeded = append(result.Succeeded, p)
	}
	return result
}
