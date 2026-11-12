package files

import "os/exec"

// Open opens path with its default application. It runs macOS open(1) and waits
// for it to finish.
func Open(path string) error {
	return exec.Command("/usr/bin/open", path).Run()
}

// OpenWith opens path with the app called appName, such as "Terminal" or
// "Visual Studio Code". It runs macOS open(1) and waits for it to finish.
func OpenWith(appName, path string) error {
	return exec.Command("/usr/bin/open", openArgs(appName, path)...).Run()
}

// openArgs returns the arguments that make open(1) use appName for path.
func openArgs(appName, path string) []string {
	return []string{"-a", appName, path}
}
