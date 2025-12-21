package env

import (
	"os"
	"path/filepath"
	"strings"
)

// LocalBin returns ~/.local/bin.
func LocalBin() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// EnsureLocalBinInPath ensures ~/.local/bin is present in PATH for the current process.
//
// This does NOT modify the parent shell's environment. It only affects the current
// nvimwiz process (and children it spawns). That is still enough to:
//   - refresh sysinfo immediately after install (no relaunch needed)
//   - find newly installed tools when running follow-up steps (e.g. :Lazy sync)
func EnsureLocalBinInPath() (changed bool, localBin string, err error) {
	localBin, err = LocalBin()
	if err != nil {
		return false, "", err
	}

	path := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	parts := strings.Split(path, sep)
	for _, p := range parts {
		if p == localBin {
			return false, localBin, nil
		}
	}

	// Prepend so user-local tools win over system ones.
	_ = os.Setenv("PATH", localBin+sep+path)
	return true, localBin, nil
}

// InPath reports whether the given directory is included in PATH.
func InPath(dir string) bool {
	path := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	for _, p := range strings.Split(path, sep) {
		if p == dir {
			return true
		}
	}
	return false
}
