package install

import (
	"os/exec"
	"strings"
)

// normalizeVersion turns common version strings into something comparable.
//
// Examples:
//   - "v0.11.5" -> "0.11.5"
//   - "0.11.5"  -> "0.11.5"
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

// installedCommandVersion tries to resolve a command in PATH and parse its version.
// It returns the normalized version, resolved path, and ok.
func installedCommandVersion(command string, args ...string) (version string, path string, ok bool) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", "", false
	}

	cmd := exec.Command(path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", path, false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return normalizeVersion(fields[1]), path, true
		}
		break
	}

	return "", path, false
}
