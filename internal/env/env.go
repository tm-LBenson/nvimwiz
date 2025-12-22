package env

import (
	"os"
	"path/filepath"
	"strings"
)

func LocalBin() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "bin"), nil
}

func EnsureLocalBinInPath() (bool, string, error) {
	lb, err := LocalBin()
	if err != nil {
		return false, "", err
	}
	sep := string(os.PathListSeparator)
	parts := strings.Split(os.Getenv("PATH"), sep)
	for _, p := range parts {
		if p == lb {
			return false, lb, nil
		}
	}
	os.Setenv("PATH", lb+sep+os.Getenv("PATH"))
	return true, lb, nil
}
